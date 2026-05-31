package sniproxy

import (
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"shanhu.io/std/errcode"
)

type endpointAddr struct {
	addr string
}

func (*endpointAddr) Network() string { return "sniproxy" }

func (a *endpointAddr) String() string { return a.addr }

// Endpoint is an endpoint that implements the listener.
type Endpoint struct {
	conn      *websocket.Conn
	addr      string
	server    *endpointServer
	serveErr  error
	serveDone chan struct{}
	incoming  chan net.Conn
	closed    chan struct{}
	closeOnce sync.Once
}

// newEndpoint creates a new endpoint based on the given websocket connection.
func newEndpoint(
	conn *websocket.Conn, d *websocketDialer, opt *Options,
) *Endpoint {
	ep := &Endpoint{
		conn:      conn,
		addr:      d.address(),
		server:    newEndpointServer(conn, d, opt),
		serveDone: make(chan struct{}),
		incoming:  make(chan net.Conn, 10),
		closed:    make(chan struct{}),
	}
	ep.server.setAccept(ep.sendAccept)
	go ep.serve()
	return ep
}

var (
	errEndpointClosed = errcode.Internalf("tunnel closed")
	errAcceptTimeout  = errcode.TimeOutf("accept timeout")
)

func (p *Endpoint) sendAccept(conn net.Conn) error {
	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
		return errAcceptTimeout
	case p.incoming <- conn:
		return nil
	case <-p.closed:
		return errEndpointClosed
	}
}

func (p *Endpoint) serve() {
	p.serveErr = p.server.serve()
	close(p.serveDone)
}

// Accept accepts a connection from the tunnel.
func (p *Endpoint) Accept() (net.Conn, error) {
	select {
	case conn := <-p.incoming:
		return conn, nil
	case <-p.serveDone:
		if p.serveErr == nil {
			return nil, io.EOF // connection closed
		}
		return nil, p.serveErr
	case <-p.closed:
		return nil, errEndpointClosed
	}
}

// Close closes the endpoint. It closes the tunnel, so all accepted, unclosed
// connections are also lost.
func (p *Endpoint) Close() error {
	err := errAlreadyClosed
	p.closeOnce.Do(func() {
		first := new(firstErr)
		first.set(p.server.SendShutdownHint())

		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		select {
		case <-timer.C:
			first.set(errcode.TimeOutf("graceful close timeout"))
		case <-p.serveDone:
			first.set(p.serveErr)
		}

		close(p.closed)
		first.set(p.conn.Close())

		err = first.get()
	})
	return err
}

// Addr returns the network address of the endpoint.
func (p *Endpoint) Addr() net.Addr {
	return &endpointAddr{addr: p.addr}
}
