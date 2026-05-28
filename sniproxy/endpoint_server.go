package sniproxy

import (
	"context"
	"io"
	"log"
	mrand "math/rand"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"shanhu.io/g/errcode"
	"shanhu.io/g/rand"
)

type endpointServer struct {
	conn    *websocket.Conn
	dialer  *websocketDialer
	options *Options
	rand    *mrand.Rand

	writeMu    sync.Mutex
	callWait   sync.WaitGroup
	sessionID  *sessionID
	acceptConn func(net.Conn) error

	conns *connections
}

func newEndpointServer(
	c *websocket.Conn, d *websocketDialer, opt *Options,
) *endpointServer {
	return &endpointServer{
		conn:      c,
		dialer:    d,
		options:   opt,
		rand:      rand.New(),
		sessionID: newSessionID(),
		conns:     newConnections(),
	}
}

func (s *endpointServer) setAccept(f func(net.Conn) error) {
	s.acceptConn = f
}

func (s *endpointServer) SendShutdownHint() error {
	x := &endpointExchange{t: msgShutdownHint}
	return s.writeResp(x)
}

func (s *endpointServer) writeResp(x *endpointExchange) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return writeExchangeResp(s.conn, x)
}

func (s *endpointServer) startCall(r io.Reader) (*endpointExchange, error) {
	dec := newDecoder(r)
	id := dec.u64()
	t := dec.u8()
	if dec.hasErr() {
		return nil, dec.Err()
	}

	x := &endpointExchange{id: id, t: t}
	if req, ok := newRequestMessage(t); ok {
		x.req = req
	} else {
		dec.end()
		return x, nil
	}
	if x.req != nil {
		x.req.decodeFrom(dec)
	}
	dec.end()

	if dec.hasErr() {
		return nil, dec.Err()
	}
	return x, nil
}

func (s *endpointServer) serveCall(x *endpointExchange) error {
	switch x.t {
	case msgHello:
		x.resp = s.handleHello(x.req.(*helloRequest))
	case msgDial:
		x.resp = s.handleDial(x.req.(*dialRequest))
	case msgDialSide:
		x.resp = s.handleDialSide(x.req.(*dialSideRequest))
	case msgDialSide2:
		x.resp = s.handleDialSide2(x.req.(*dialSide2Request))
	case msgRead:
		x.resp = s.handleRead(x.req.(*readRequest))
	case msgWrite:
		x.resp = s.handleWrite(x.req.(*writeRequest))
	case msgClose:
		x.resp = s.handleClose(x.req.(*closeRequest))
	case msgShutdown:
		// No need for special handling here.
	default:
		x.errcode = errUnknownType
	}
	return s.writeResp(x)
}

func (s *endpointServer) handleHello(req *helloRequest) *helloResponse {
	return &helloResponse{msg: req.msg}
}

func (s *endpointServer) sideConn(tok string, k *sessionKey, addr string) (
	net.Conn, error,
) {
	bg := context.Background()
	ctx, cancel := context.WithTimeout(bg, 5*time.Second)
	defer cancel()

	conn, err := s.dialer.dialSide(ctx, tok, k)
	if err != nil {
		return nil, err
	}
	return newSideConn(conn, addr), nil
}

func (s *endpointServer) handleDialSide2(
	req *dialSide2Request,
) *dialResponse {
	if !s.options.Siding {
		return &dialResponse{
			err: newRemoteErrString(errAccept, "needs legacy dialing"),
		}
	}
	if s.acceptConn == nil {
		return &dialResponse{err: remoteErrNotAccepting}
	}

	k := &sessionKey{ID: req.session, Key: req.key}
	id := s.sessionID.next() // server side id, just for accounting.
	resp := &dialResponse{session: id}
	conn, err := s.sideConn(req.token, k, req.tcpAddr)
	if err != nil {
		resp.err = newRemoteErr(errAccept, err)
		return resp
	}
	if err := s.acceptConn(conn); err != nil {
		conn.Close()
		resp.err = newRemoteErr(errAccept, err)
		return resp
	}
	// Can transfer data now.
	return resp
}

func (s *endpointServer) handleDialSide(req *dialSideRequest) *dialResponse {
	req2 := &dialSide2Request{
		session: req.session,
		key:     req.key,
		token:   req.token,
	}
	return s.handleDialSide2(req2)
}

func (s *endpointServer) handleDial(req *dialRequest) *dialResponse {
	if s.options.Siding {
		return &dialResponse{err: remoteErrSiding}
	}
	if s.acceptConn == nil {
		return &dialResponse{err: remoteErrNotAccepting}
	}

	id := s.sessionID.next()
	conn := newConnection(id)
	defer func() {
		if conn != nil {
			conn.cleanup()
		}
	}()

	if err := s.acceptConn(conn.forServer()); err != nil {
		rerr := newRemoteErr(errAccept, err)
		return &dialResponse{session: id, err: rerr}
	}

	if err := s.conns.add(conn); err != nil {
		rerr := newRemoteErr(errAccept, err)
		return &dialResponse{session: id, err: rerr}
	}
	conn = nil // belongs to s.conns now.

	return &dialResponse{session: id}
}

func (s *endpointServer) findSession(id uint64) (*connection, *remoteErr) {
	c, err := s.conns.get(id)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, newRemoteErr(errSessionNotFound, err)
		}
		return nil, newRemoteErr(errInternal, err)
	}
	return c, nil
}

func (s *endpointServer) handleRead(req *readRequest) *readResponse {
	if s.options.Siding {
		return &readResponse{err: remoteErrSiding}
	}
	conn, rerr := s.findSession(req.session)
	if rerr != nil {
		return &readResponse{err: rerr}
	}
	buf := make([]byte, req.maxRead)
	n, err := conn.Read(buf)
	resp := &readResponse{bytes: buf[:n]}
	if err != nil {
		if err == io.EOF {
			resp.err = newRemoteErrString(errEOF, "eof")
		} else {
			resp.err = newRemoteErr(errRead, err)
		}
	}
	return resp
}

func (s *endpointServer) handleWrite(req *writeRequest) *writeResponse {
	if s.options.Siding {
		return &writeResponse{err: remoteErrSiding}
	}
	conn, rerr := s.findSession(req.session)
	if rerr != nil {
		return &writeResponse{err: rerr}
	}
	n, err := conn.Write(req.bytes)
	resp := &writeResponse{written: n}
	if err != nil {
		resp.err = newRemoteErr(errWrite, err)
	}
	return resp
}

func (s *endpointServer) handleClose(req *closeRequest) *closeResponse {
	if s.options.Siding {
		return &closeResponse{err: remoteErrSiding}
	}
	conn, rerr := s.findSession(req.session)
	if rerr != nil {
		return &closeResponse{err: rerr}
	}
	resp := &closeResponse{}
	err := conn.Close()
	if err != nil {
		resp.err = newRemoteErr(errClose, err)
	}
	if err := s.conns.remove(req.session); err != nil {
		resp.err = newRemoteErr(errClose, err)
	}
	return resp
}

func (s *endpointServer) cleanup() {
	conns := s.conns.shutdown()
	for _, conn := range conns {
		conn.cleanup()
	}
}

func (s *endpointServer) serve() error {
	defer func() {
		s.cleanup()       // Closes all connections.
		s.callWait.Wait() // Join all pending calls.
	}()

	for {
		typ, r, err := s.conn.NextReader()
		if err != nil {
			return err
		}
		if typ != websocket.BinaryMessage {
			return errcode.InvalidArgf("invalid message type: %d", typ)
		}

		ex, err := s.startCall(r)
		if err != nil {
			return errcode.Internalf("decode message: %s", err)
		}
		s.callWait.Add(1)
		go func(ex *endpointExchange) {
			defer s.callWait.Done()
			if err := s.serveCall(ex); err != nil {
				log.Printf("serve call: %s", err)
			}
		}(ex)

		if ex.t == msgShutdown {
			return nil // Exit gracefully. Stop accepting other calls.
		}
	}
}
