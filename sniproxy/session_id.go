package sniproxy

import (
	"sync"
)

type sessionID struct {
	id uint64
	mu sync.Mutex
}

func newSessionID() *sessionID {
	return &sessionID{}
}

func (id *sessionID) next() uint64 {
	id.mu.Lock()
	defer id.mu.Unlock()
	ret := id.id
	id.id++
	return ret
}
