package sniproxy

import (
	"net"
)

type connection struct {
	id         uint64
	net.Conn   // clientConn
	serverConn net.Conn
}

func newConnection(id uint64) *connection {
	client, server := net.Pipe()

	return &connection{
		id:         id,
		Conn:       client,
		serverConn: server,
	}
}

func (c *connection) forServer() net.Conn {
	return c.serverConn
}

func (c *connection) cleanup() {
	c.Conn.Close()
	c.serverConn.Close()
}

func (c *connection) session() uint64 { return c.id }
