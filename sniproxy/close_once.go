package sniproxy

import (
	"io"
	"sync"
)

type closerOnce struct {
	io.Closer
	once sync.Once
	err  error
}

func (c *closerOnce) Close() error {
	c.once.Do(func() {
		c.err = c.Closer.Close()
	})
	return c.err
}
