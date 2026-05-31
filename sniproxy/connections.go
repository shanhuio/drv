package sniproxy

import (
	"fmt"
	"sync"

	"shanhu.io/std/errcode"
)

type connections struct {
	closed bool
	m      map[uint64]*connection
	mu     sync.Mutex
}

func newConnections() *connections {
	return &connections{
		m: make(map[uint64]*connection),
	}
}

func (cs *connections) add(c *connection) error {
	id := c.session()

	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.closed {
		return errAlreadyShutdown
	}

	if _, found := cs.m[id]; found {
		return fmt.Errorf("session id conflict: %d", id)
	}
	cs.m[id] = c
	return nil
}

func (cs *connections) get(id uint64) (*connection, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if cs.closed {
		return nil, errAlreadyShutdown
	}
	c, ok := cs.m[id]
	if !ok {
		return nil, errcode.NotFoundf("session not found: %d", id)
	}
	return c, nil
}

func (cs *connections) remove(id uint64) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.closed {
		return errAlreadyShutdown
	}

	if _, found := cs.m[id]; !found {
		return errcode.NotFoundf("session not found: %d", id)
	}
	delete(cs.m, id)
	return nil
}

func (cs *connections) shutdown() map[uint64]*connection {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.closed {
		return nil
	}
	cs.closed = true
	return cs.m
}
