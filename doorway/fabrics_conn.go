package doorway

import (
	"net"

	"shanhu.io/drv/sniproxy"
)

type fabricsAddr struct {
	addr string
}

func (*fabricsAddr) Network() string  { return "tcp" }
func (a *fabricsAddr) String() string { return "|" + a.addr }

type fabricsConn struct {
	net.Conn
}

func (c *fabricsConn) Addr() net.Addr {
	return &fabricsAddr{addr: c.Conn.RemoteAddr().String()}
}

type fabricsListener struct {
	*sniproxy.Endpoint
}

func (l *fabricsListener) Accept() (net.Conn, error) {
	conn, err := l.Endpoint.Accept()
	if err != nil {
		return nil, err
	}
	return &fabricsConn{Conn: conn}, nil
}
