package sniproxy

import (
	"context"
	"fmt"
	"net"
	"time"
)

type tunnelAddr struct {
	session uint64
}

func (a *tunnelAddr) Network() string { return "tunnel" }
func (a *tunnelAddr) String() string  { return fmt.Sprintf("#%d", a.session) }

// tunnel is an established of a session. On one end, it is an endpoint that is
// serving via a websocket connection, on the other end, it provides a
// io.ReadWriteCloser interface for the incoming connection.
type tunnel struct {
	ctx     context.Context
	tr      *transport
	session uint64
}

func newTunnel(tr *transport, session uint64) *tunnel {
	return &tunnel{
		tr:      tr,
		session: session,
		ctx:     context.TODO(),
	}
}

func (t *tunnel) Write(bs []byte) (int, error) {
	req := &writeRequest{
		session: t.session,
		bytes:   bs,
	}
	resp := new(writeResponse)

	if err := t.tr.call(t.ctx, msgWrite, req, resp); err != nil {
		return 0, err
	}
	return resp.written, resp.err.toError()
}

func (t *tunnel) Read(buf []byte) (int, error) {
	req := &readRequest{
		session: t.session,
		maxRead: len(buf),
	}
	resp := &readResponse{bytes: buf}
	if err := t.tr.call(t.ctx, msgRead, req, resp); err != nil {
		return 0, err
	}
	return len(resp.bytes), resp.err.toError()
}

func (t *tunnel) Close() error {
	req := &closeRequest{session: t.session}
	resp := new(closeResponse)
	if err := t.tr.call(t.ctx, msgClose, req, resp); err != nil {
		return err
	}
	return resp.err.toError()
}

func (t *tunnel) LocalAddr() net.Addr  { return &tunnelAddr{t.session} }
func (t *tunnel) RemoteAddr() net.Addr { return &tunnelAddr{t.session} }

func (t *tunnel) SetDeadline(d time.Time) error {
	if err := t.SetReadDeadline(d); err != nil {
		return err
	}
	return t.SetWriteDeadline(d)
}

func (t *tunnel) SetReadDeadline(d time.Time) error {
	// TODO(h8liu): implement this
	return nil
}
func (t *tunnel) SetWriteDeadline(d time.Time) error {
	// TODO(h8liu): implement this
	return nil
}
