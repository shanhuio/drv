package sniproxy

import (
	"github.com/gorilla/websocket"
)

// callExchange captures the state of a single call at the client
// transport.
type callExchange struct {
	typ  uint8
	id   uint64
	req  encoderTo
	resp decoderFrom
	err  error
	done func()
}

func newCallExchange(c *transportCall) *callExchange {
	x := &callExchange{
		typ:  c.typ,
		req:  c.req,
		resp: c.resp,
	}
	x.done = func() { c.done(x.err) }
	return x
}

// pendingFetch fetches a pending callExchange based on the call id.
// if no call is found
type pendingFetch struct {
	id   uint64
	call chan *callExchange
}

func sendExchangeReq(conn *websocket.Conn, c *callExchange) error {
	w, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	defer w.Close()

	enc := newEncoder(w)
	enc.u64(c.id)
	enc.u8(c.typ)

	if c.req != nil {
		c.req.encodeTo(enc)
	}
	if enc.hasErr() {
		return enc.Err()
	}
	return w.Close()
}
