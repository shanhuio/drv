package sniproxy

import (
	"context"
	"net"
	"sync"

	"shanhu.io/g/errcode"
)

type connMailBox struct {
	key       *sessionKey
	office    *connMailOffice
	ch        chan net.Conn
	closed    chan struct{}
	closeOnce sync.Once
}

func (b *connMailBox) Close() error {
	b.closeOnce.Do(func() { close(b.closed) })
	return nil
}

func (b *connMailBox) cleanUp() {
	b.Close()
	b.office.remove(b.key)
}

func (b *connMailBox) receive(ctx context.Context) (net.Conn, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-b.closed:
		return nil, errcode.TimeOutf("closed")
	case conn := <-b.ch:
		return conn, nil
	}
}

func (b *connMailBox) match(k *sessionKey) bool {
	return b.key.ID == k.ID && b.key.Key == k.Key
}

func (b *connMailBox) deliver(conn net.Conn) {
	select {
	case b.ch <- conn:
	default:
	}
}

type connMailOffice struct {
	mu sync.Mutex
	m  map[uint64]*connMailBox
}

func newConnMailOffice() *connMailOffice {
	return &connMailOffice{
		m: make(map[uint64]*connMailBox),
	}
}

func (o *connMailOffice) newBox(k *sessionKey) *connMailBox {
	o.mu.Lock()
	defer o.mu.Unlock()

	if cur, ok := o.m[k.ID]; ok {
		cur.Close() // close the current pending one.
	}
	b := &connMailBox{
		key:    k,
		office: o,
		ch:     make(chan net.Conn, 1),
		closed: make(chan struct{}),
	}
	o.m[k.ID] = b
	return b
}

func (o *connMailOffice) deliver(k *sessionKey, conn net.Conn) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	box, ok := o.m[k.ID]
	if !ok {
		return errcode.NotFoundf("session not found")
	}
	if !box.match(k) {
		return errcode.InvalidArgf("key mismatch")
	}
	box.deliver(conn)
	return nil
}

func (o *connMailOffice) remove(k *sessionKey) {
	o.mu.Lock()
	defer o.mu.Unlock()
	box, ok := o.m[k.ID]
	if !ok {
		return
	}
	if box.match(k) {
		delete(o.m, k.ID)
	}
}
