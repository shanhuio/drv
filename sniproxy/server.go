package sniproxy

import (
	"context"
	"log"
	"net"
	"sync"

	"github.com/gorilla/websocket"
	"shanhu.io/g/aries"
	"shanhu.io/std/errcode"
)

// Dest specifies a destination.
type Dest struct {
	Name       string
	Home       bool // Use DialHome.
	ForwardTCP string
}

// ServerConfig contains configuration of an SNI based proxy server.
type ServerConfig struct {
	// Lookup looks for the user ID for a particular domain.
	Lookup func(domain string) (*Dest, error)

	// DialHome provides a dialer for dialing home for endpoint name "~".
	DialHome func(ctx context.Context) (net.Conn, error)

	// DialForward provides a dialer for dialing a domain for endpoint
	// that is fowarding to a TCP. If this is not provided,
	// a default network TCP dailing will be used.
	DialForward func(ctx context.Context, fwd string) (net.Conn, error)

	// SideToken gets a token for side connections.
	SideToken func(user string) (string, error)

	// OnConnect is called when a new endpoint connects. It returns
	// a session ID. This callback function is optional.
	OnConnect func(user string) int64

	// OnDisconnect is called when a new endpoint disconnects. It is
	// called with the user's name and the session ID got from
	// OnConnect. If OnConnect is not set, session is always 0.
	OnDisconnect func(user string, session int64)
}

// Server is an SNI based TCP proxy server that can serve over websocket.
type Server struct {
	upgrader *websocket.Upgrader
	proxy    *proxy

	lookup      func(domain string) (*Dest, error)
	dialHome    func(ctx context.Context) (net.Conn, error)
	dialForward func(ctx context.Context, domain string) (net.Conn, error)

	mu        sync.Mutex
	endpoints map[string]*endpointClient

	callback func(user string, ep *endpointClient)

	sideToken func(user string) (string, error)

	onConnect    func(user string) int64
	onDisconnect func(user string, session int64)
}

// NewServer creates a new server that can accept endpoint providing
// websocket connections.
func NewServer(config *ServerConfig) *Server {
	s := &Server{
		upgrader: &websocket.Upgrader{
			ReadBufferSize:  DefaultReadBufferSize,
			WriteBufferSize: DefaultWriteBufferSize,
		},
		lookup:       config.Lookup,
		dialHome:     config.DialHome,
		dialForward:  config.DialForward,
		endpoints:    make(map[string]*endpointClient),
		sideToken:    config.SideToken,
		onConnect:    config.OnConnect,
		onDisconnect: config.OnDisconnect,
	}
	s.proxy = newProxy(s)
	return s
}

func (s *Server) setEndpointCallback(f func(u string, ep *endpointClient)) {
	s.callback = f
}

// ServeFront starts accepting connections from the configured net.Listener and
// route incoming connections to connected endpoints.
func (s *Server) ServeFront(ctx context.Context, lis net.Listener) error {
	return s.proxy.serve(ctx, lis)
}

func (s *Server) endpoint(name string) (*endpointClient, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.endpoints[name]
	if !ok {
		return nil, errcode.NotFoundf("not found")
	}
	return c, nil
}

func endpointNotFoundError(domain string) error {
	return errcode.NotFoundf("endpoint for %q not found", domain)
}

var defaultDialer = &net.Dialer{}

func (s *Server) dial(
	ctx context.Context, hello *TLSHelloInfo, asAddr string,
) (net.Conn, error) {
	if s.lookup == nil {
		return nil, errcode.Internalf("server not accepting")
	}
	domain := hello.ServerName
	dest, err := s.lookup(domain)
	if err != nil {
		return nil, err
	}

	if dest.Home {
		if s.dialHome == nil {
			return nil, endpointNotFoundError(domain)
		}
		return s.dialHome(ctx)
	} else if fwd := dest.ForwardTCP; fwd != "" {
		if s.dialForward == nil {
			return defaultDialer.DialContext(ctx, "tcp", fwd)
		}
		return s.dialForward(ctx, fwd)
	}

	ep, err := s.endpoint(dest.Name)
	if err != nil {
		return nil, errcode.Annotatef(err, "endpoint for %q", domain)
	}
	return ep.Dial(ctx, asAddr)
}

func checkEndpointName(name string) error {
	if name == "" {
		return errcode.InvalidArgf("invalid endpoint name: %q", name)
	}
	return nil
}

func (s *Server) unmap(name string, ep *endpointClient) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.endpoints[name] == ep {
		log.Printf("unmap endpoint %q", name)
		delete(s.endpoints, name)
	}
}

func (s *Server) upgrade(c *aries.C, name string, opt *Options) (
	*endpointClient, error,
) {
	conn, err := s.upgrader.Upgrade(c.Resp, c.Req, nil)
	if err != nil {
		return nil, err
	}
	ep := newEndpointClient(conn, opt)
	if opt.Siding {
		ep.setToken(func() (string, error) {
			if s.sideToken == nil {
				return "", nil
			}
			return s.sideToken(name)
		})
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if old, found := s.endpoints[name]; found {
		log.Printf("kick off old endpoint for %q", name)
		delete(s.endpoints, name)

		// Send background to kill gracefully.
		go func(name string, old *endpointClient) {
			if err := old.Close(); err != nil {
				log.Printf("close old endpoint %s: %s", name, err)
			} else {
				log.Printf("old endpoint %s closed", name)
			}
		}(name, old)
	}
	log.Printf("map endpoint %q", name)
	s.endpoints[name] = ep
	return ep, nil
}

// ServeBackName serves an incoming proxy connection via websocket using the
// given endpoint name.
func (s *Server) ServeBackName(c *aries.C, name string) error {
	if err := checkEndpointName(name); err != nil {
		return err
	}

	query := c.Req.URL.Query()
	if side := query.Get("side"); side != "" {
		k, err := decodeSessionKey(side)
		if err != nil {
			return errcode.InvalidArgf("invalid session: %s", err)
		}
		return s.serveBackSide(c, name, k)
	}
	opt, err := optionsFromQuery(query)
	if err != nil {
		return err
	}

	// Upgrade into an endpoint client.
	ep, err := s.upgrade(c, name, opt)
	if err != nil {
		return err
	}
	defer func() {
		s.unmap(name, ep)
		ep.Close()
	}()

	if s.callback != nil {
		s.callback(name, ep)
	}
	var session int64
	if s.onConnect != nil {
		session = s.onConnect(name)
	}
	defer func() {
		if s.onDisconnect != nil {
			s.onDisconnect(name, session)
		}
	}()

	if err := ep.serve(); err != nil {
		log.Printf("serve endpoint: %s", err)
	}
	return nil
}

// ServeBack serves an incoming proxy connection via websocket.
// It uses c.User as the endpoint name
func (s *Server) ServeBack(c *aries.C) error {
	return s.ServeBackName(c, c.User)
}

func (s *Server) serveBackSide(
	c *aries.C, name string, k *sessionKey,
) error {
	ep, err := s.endpoint(name)
	if err != nil {
		return errcode.Annotate(err, "find endpoint")
	}

	wsConn, err := s.upgrader.Upgrade(c.Resp, c.Req, nil)
	if err != nil {
		return err
	}
	conn := newSideConn(wsConn, "")
	defer conn.Close()
	if err := ep.deliverSideConn(k, conn); err != nil {
		return errcode.Annotate(err, "register session")
	}

	// Now the connection is handled to the client, and we just wait for the
	// client to close this (or the http connection is lost).
	//
	// Note that we cannot close the connection when Read() or Write()
	// encounters an error, as it might be a timeout (via one of the
	// SetDeadline() methods), or it might be read reaches EOF, but still
	// has more stuff to write.
	conn.wait(c.Context)

	return nil
}
