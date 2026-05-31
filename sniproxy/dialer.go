package sniproxy

import (
	"context"
	"net"

	"shanhu.io/std/errcode"
)

type dialer interface {
	dial(ctx context.Context, hello *TLSHelloInfo, asAddr string) (
		net.Conn, error,
	)
}

type tcpDialer struct {
	raddrs map[string]*net.TCPAddr
}

func (d *tcpDialer) dial(_ context.Context, hello *TLSHelloInfo, _ string) (
	net.Conn, error,
) {
	domain := hello.ServerName
	raddr, ok := d.raddrs[domain]
	if !ok {
		return nil, errcode.NotFoundf("no connection for domain %q", domain)
	}

	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}
	return conn, nil
}
