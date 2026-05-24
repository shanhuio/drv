package doorway

import (
	"net"
)

const (
	tagTCP     = "TCP"
	tagFabrics = "fabrics"
)

type tagConnListener interface {
	net.Listener
	acceptTag() (*tagConn, error)
}

type tagListener struct {
	net.Listener
	tag string
}

func newTagListener(lis net.Listener, tag string) *tagListener {
	return &tagListener{
		Listener: lis,
		tag:      tag,
	}
}

func (l *tagListener) Accept() (net.Conn, error) {
	conn, err := l.acceptTag()
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (l *tagListener) acceptTag() (*tagConn, error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &tagConn{Conn: conn, tag: l.tag}, nil
}

type tagConn struct {
	net.Conn
	tag string
}
