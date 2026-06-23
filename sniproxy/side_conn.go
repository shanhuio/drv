package sniproxy

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"shanhu.io/std/errcode"
)

type sideConnAddr struct {
	addr string
}

func newSideConnAddr(addr string) *sideConnAddr {
	return &sideConnAddr{addr: addr}
}

func (a *sideConnAddr) Network() string { return "tcp" }
func (a *sideConnAddr) String() string  { return a.addr }

type sideConn struct {
	*websocket.Conn
	addr net.Addr

	writeMu     sync.Mutex
	writeClosed bool

	readMu    sync.Mutex
	curReader io.Reader

	closeOnce sync.Once
	closed    chan struct{}
}

func newSideConn(conn *websocket.Conn, addr string) *sideConn {
	c := &sideConn{
		Conn:   conn,
		closed: make(chan struct{}),
	}
	if addr == "" {
		c.addr = conn.RemoteAddr()
	} else {
		c.addr = newSideConnAddr(addr)
	}
	return c
}

func (c *sideConn) RemoteAddr() net.Addr { return c.addr }

func (c *sideConn) nextReader() error {
	t, r, err := c.Conn.NextReader()
	if err != nil {
		if websocket.IsCloseError(err) {
			closeErr := err.(*websocket.CloseError)
			if closeErr.Code == websocket.CloseNormalClosure {
				return io.EOF
			}
		}
		return err
	}
	if t == websocket.TextMessage { // text message is EOF.
		return io.EOF
	}
	c.curReader = r
	return nil
}

func (c *sideConn) Read(buf []byte) (int, error) {
	c.readMu.Lock()
	defer c.readMu.Unlock()

	if c.curReader == nil {
		if err := c.nextReader(); err != nil {
			return 0, err
		}
	}

	for {
		n, err := c.curReader.Read(buf)
		if err != io.EOF { // not the end of reader or has error
			return n, err
		}

		// err is EOF: end of current reader
		c.curReader = nil // clears it
		if n > 0 {
			return n, nil
		}
		// n == 0, has not read anything.
		// needs to read again if there is more.
		if err := c.nextReader(); err != nil {
			return 0, err
		}
	}
}

func (c *sideConn) Write(buf []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	const chunk = 4096

	if c.writeClosed {
		return 0, errcode.Internalf("already closed")
	}

	n := 0
	for n < len(buf) {
		end := min(n+chunk, len(buf))
		toSend := buf[n:end]

		w, err := c.Conn.NextWriter(websocket.BinaryMessage)
		if err != nil {
			return n, err
		}
		written, err := w.Write(toSend)
		if err != nil {
			n += written
			return n, err
		}
		n += len(toSend)
		if err := w.Close(); err != nil {
			return n, err
		}
	}
	return n, nil
}

// CloseWrite closes the writer to indicates an EOF.
// This is also called when calling Close().
func (c *sideConn) CloseWrite() error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if c.writeClosed {
		return nil
	}
	c.writeClosed = true

	w, err := c.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte("EOF")); err != nil {
		return err
	}
	return w.Close()
}

func (c *sideConn) SetDeadline(d time.Time) error {
	rerr := c.Conn.SetReadDeadline(d)
	werr := c.Conn.SetWriteDeadline(d)
	if rerr != nil {
		return rerr
	}
	return werr
}

func (c *sideConn) Close() error {
	deadline := time.Now().Add(3 * time.Second)
	c.SetDeadline(deadline)

	werr := c.CloseWrite() // sends graceful EOF.
	err := c.Conn.Close()
	c.closeOnce.Do(func() { close(c.closed) })
	if werr != nil {
		return werr
	}
	return err
}

func (c *sideConn) wait(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-c.closed:
	}
}
