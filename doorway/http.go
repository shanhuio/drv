package doorway

import (
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"shanhu.io/g/aries"
	"shanhu.io/g/aries/redhttp"
)

// HTTPServerConfig is the configuration for the http redirection service.
type HTTPServerConfig struct {
	// Address to listen on.
	Addr string

	// When the host is an IP or a ".local" address, forward to this service.
	LocalMapping string
}

type httpServer struct {
	localMapping string
	addr         string

	proxy *httputil.ReverseProxy
}

func newHTTPServer(config *HTTPServerConfig) *httpServer {
	s := &httpServer{
		addr:         config.Addr,
		localMapping: config.LocalMapping,
	}
	s.proxy = &httputil.ReverseProxy{Director: s.localDirector}
	return s
}

func (s *httpServer) localDirector(req *http.Request) {
	req.Header.Set("X-Forwarded-Proto", "http")

	if s.localMapping == "" {
		req.URL = sinkURL
		return
	}
	forwardToHTTP(req, s.localMapping)
}

func (s *httpServer) Serve(c *aries.C) error {
	host := strings.TrimSuffix(c.Req.Host, ".")
	if net.ParseIP(host) != nil || strings.HasSuffix(host, ".local") {
		s.proxy.ServeHTTP(c.Resp, c.Req)
		return nil
	}
	return redhttp.Redirect(c)
}

func (s *httpServer) Addr() string {
	if s.addr == "" {
		return ":8080"
	}
	return s.addr
}

func runHTTPServer(h *httpServer) {
	addr := h.Addr()
	log.Printf("starts http on %q", addr)

	for {
		s := &http.Server{
			Addr:    addr,
			Handler: aries.Serve(h),
		}
		if err := s.ListenAndServe(); err != nil {
			log.Print(err)
		}
		s.Close()
		time.Sleep(time.Second)
	}
}
