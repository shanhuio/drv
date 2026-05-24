package doorway

import (
	"net"

	"shanhu.io/g/errcode"
)

type localListenConfig struct {
	// listener is the listener to listen on.
	listener net.Listener

	// addr is the TCP address to listen on. This is used when listener is nil.
	addr string
}

type listenConfig struct {
	// local listens on local TCP address or a specific provided listener.
	local *localListenConfig

	// fabrics is for listening on a fabrics connection.
	fabrics *fabricsConfig
}

func listenLocal(c *localListenConfig) (*tagListener, error) {
	if c.listener != nil {
		return newTagListener(c.listener, tagTCP), nil
	}
	if c.addr == "" {
		return nil, errcode.InvalidArgf("listen address missing")
	}
	tcp, err := net.Listen("tcp", c.addr)
	if err != nil {
		return nil, errcode.Annotate(err, "listen local")
	}
	return newTagListener(tcp, tagTCP), nil
}

func listen(ctx C, c *listenConfig) (tagConnListener, error) {
	if c.local == nil && c.fabrics == nil {
		return nil, errcode.InvalidArgf(
			"must listen either at local or via fabrics",
		)
	}

	if c.fabrics == nil { // no fabrics, just listen local
		lis, err := listenLocal(c.local)
		if err != nil {
			return nil, err
		}
		return lis, nil
	}

	if c.local == nil {
		lis, err := listenFabrics(ctx, c.fabrics)
		if err != nil {
			return nil, err
		}
		return lis, nil
	}

	// Dual listener.
	local, err := listenLocal(c.local)
	if err != nil {
		return nil, err
	}
	fabLis, err := listenFabrics(ctx, c.fabrics)
	if err != nil {
		local.Close()
		return nil, errcode.Annotate(err, "listen fabrics")
	}
	return newTunnelListener(local, fabLis), nil
}
