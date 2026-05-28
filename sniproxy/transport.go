package sniproxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// transport is effectively an RPC client that runs over a websocket
// connection. It is capable of performing multiple concurrent calls at the
// same time.
type transport struct {
	conn         *websocket.Conn
	calls        chan *callExchange
	pendingFetch chan *pendingFetch

	shutdownSignal chan struct{}
	shutdownOnce   sync.Once
	serveDone      chan struct{}
}

func newTransport(conn *websocket.Conn, opt *Options) *transport {
	return &transport{
		conn:           conn,
		calls:          make(chan *callExchange, 128),
		pendingFetch:   make(chan *pendingFetch, 5),
		shutdownSignal: make(chan struct{}),
		serveDone:      make(chan struct{}),
	}
}

var errTooLong = errors.New("pending call for too long")

func (tr *transport) startShutdown() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := tr.shutdown(ctx); err != nil {
			log.Printf("transport shutdown: %s", err)
		}
	}()
}

func (tr *transport) handleMessage(r io.Reader) error {
	dec := newDecoder(r)
	id := dec.u64()
	typ := dec.u8()
	errcode := dec.u8()
	if dec.hasErr() {
		err := dec.Err()
		if err == io.ErrUnexpectedEOF {
			log.Printf("small packet received: %d bytes", dec.count())
			return nil
		}
		return err
	}
	if errcode != 0 {
		dec.end()
		return fmt.Errorf("got error: %d", errcode)
	}
	if typ == msgShutdownHint {
		tr.startShutdown()
		return nil
	}

	ch := make(chan *callExchange)
	tr.pendingFetch <- &pendingFetch{id: id, call: ch}
	ex := <-ch
	if ex == nil {
		log.Printf("discard response #%d, type=%d", id, typ)
		return nil
	}
	if ex.typ != typ {
		log.Printf("response #%d, type %d!=%d", id, typ, ex.typ)
		return nil
	}
	defer ex.done()

	if ex.resp != nil {
		ex.resp.decodeFrom(dec)
	}
	if dec.hasErr() {
		ex.err = dec.Err()
		return nil
	}

	// TODO(h8liu): mark this as an invalid message?
	if _, err := io.Copy(io.Discard, r); err != nil {
		return err
	}
	if ex.typ == msgShutdown {
		return io.EOF
	}
	return nil
}

func (tr *transport) serveRead() error {
	for {
		typ, r, err := tr.conn.NextReader()
		if err != nil {
			return err
		}
		switch typ {
		case websocket.TextMessage:
			bs, err := io.ReadAll(r)
			if err != nil {
				return err
			}
			log.Printf("receive text: %s", bs)
		case websocket.BinaryMessage:
			if err := tr.handleMessage(r); err != nil {
				if err == io.EOF {
					return nil
				}
				return err
			}
		default:
			return fmt.Errorf("unknown message type: %d", typ)
		}
	}
}

func (tr *transport) send(c *callExchange) error {
	return sendExchangeReq(tr.conn, c)
}

func (tr *transport) serve() error {
	defer close(tr.serveDone)

	id := uint64(0)
	pending := make(map[uint64]*callExchange)
	defer func() {
		for _, x := range pending {
			x.err = io.ErrUnexpectedEOF
			x.done()
		}
	}()
	readErr := make(chan error, 1)

	go func() { readErr <- tr.serveRead() }()

	shutdownCalled := false
	for {
		select {
		case c := <-tr.calls:
			c.id = id
			id++ // increase the id

			if shutdownCalled {
				// After the first shutdown is called, no longer send calls
				// to the remote side.  We keep everything running only
				// because we are waiting for shutdown's returns.
				c.err = errAlreadyShutdown
				c.done()
				break
				// After shutdown is called, no new pending message will be
				// created.
			} else if c.typ == msgShutdown {
				shutdownCalled = true
			}

			if err := tr.send(c); err != nil {
				c.done()
				return err
			}

			if old, found := pending[c.id]; found {
				old.err = errTooLong
				old.done()
				delete(pending, c.id)
			}

			pending[c.id] = c
		case fetch := <-tr.pendingFetch:
			c, found := pending[fetch.id]
			if found {
				delete(pending, fetch.id)
			}
			fetch.call <- c
		case err := <-readErr:
			return err // exit when read routine closes.
		}
	}
}

func (tr *transport) hasShutdown() bool {
	select {
	case <-tr.shutdownSignal:
		return true
	default:
		return false
	}
}

func (tr *transport) asyncCall(call *transportCall) error {
	if call.typ == msgShutdown { // Special handling for shutdown.
		// Mark shutdownSignal, but only once.
		var err = errAlreadyShutdown
		tr.shutdownOnce.Do(func() {
			close(tr.shutdownSignal)
			err = nil
		})
		if err != nil {
			return err
		}
	} else if tr.hasShutdown() {
		// Every request received after shutdown request.
		// will just return errAlreadyShutdown.
		return errAlreadyShutdown
	}

	ex := newCallExchange(call)

	// Sends to the pipeline.
	ctx := call.context
	select {
	case <-ctx.Done():
		return ctx.Err()
	case tr.calls <- ex:
	}
	return nil
}

func (tr *transport) call(
	ctx context.Context, t byte, req encoderTo, resp decoderFrom,
) error {
	done := make(chan struct{})
	var err error
	c := newTransportCall(ctx, t, req, resp)
	c.done = func(e error) {
		err = e
		close(done)
	}

	if err := tr.asyncCall(c); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
	}
	return err
}

func (tr *transport) shutdown(ctx context.Context) error {
	err := tr.call(ctx, msgShutdown, nil, nil)
	select {
	case <-ctx.Done():
		if err != nil {
			return err
		}
		return ctx.Err()
	case <-tr.serveDone:
		return err
	}
}
