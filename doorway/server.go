package doorway

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"golang.org/x/crypto/acme/autocert"
	"shanhu.io/g/aries"
	"shanhu.io/std/certdelay"
	"shanhu.io/std/errcode"
)

// ServerConfig is the config for serving the reverse proxy
// server.
type ServerConfig struct {
	HostMap       map[string]string
	AutoCertCache autocert.Cache
	Home          aries.Service
	ManualCerts   map[string]*tls.Certificate
}

type server struct {
	home          aries.Service
	hostMap       hostMap
	proxy         *httputil.ReverseProxy
	autoCertCache autocert.Cache
	manualCerts   map[string]*tls.Certificate
}

func makeDefaultHome() aries.Service {
	r := aries.NewRouter()
	r.Index(aries.StringFunc("hi"))
	r.File("health", aries.StringFunc("ok"))
	return r
}

func newServer(config *ServerConfig) (*server, error) {
	s := &server{
		hostMap:       newMemHostMap(config.HostMap),
		autoCertCache: config.AutoCertCache,
		manualCerts:   config.ManualCerts,
	}

	if config.Home == nil {
		s.home = makeDefaultHome()
	} else {
		s.home = config.Home
	}

	s.proxy = &httputil.ReverseProxy{
		Director:       s.director,
		ModifyResponse: setStrictTransportSecurity,
	}
	return s, nil
}

func (s *server) Serve(c *aries.C) error {
	host := strings.TrimSuffix(c.Req.Host, ".")

	entry := s.hostMap.mapHost(host)
	if entry == nil {
		return aries.NotFound
	}

	switch entry.typ {
	default:
		return aries.NotFound
	case hostHome:
		return s.serveHome(c)
	case hostRedirect:
		u := *c.Req.URL
		u.Host = entry.host
		c.Redirect(u.String())
		return nil
	case hostProxy:
		s.proxy.ServeHTTP(c.Resp, c.Req)
		return nil
	}
}

func (s *server) serveHome(c *aries.C) error {
	return s.home.Serve(c)
}

func (s *server) director(req *http.Request) {
	// swap the scheme to http
	req.Header.Set("X-Forwarded-Proto", "https")

	host := strings.TrimSuffix(req.Host, ".")

	mapped := hostMapToProxy(s.hostMap, host)
	if mapped == "" {
		if host == "" {
			log.Println("empty host")
		} else {
			log.Printf("unexpected host: %q", host)
		}

		req.URL = sinkURL
		return
	}

	forwardToHTTP(req, mapped)
}

func setStrictTransportSecurity(resp *http.Response) error {
	resp.Header.Set(
		"Strict-Transport-Security",
		"max-age=15552000; includeSubDomains",
	)
	return nil
}

// hostPolicy determines which hosts are whitelisted for autocert.
func (s *server) hostPolicy(_ context.Context, host string) error {
	if !hostMapHas(s.hostMap, host) {
		return errcode.NotFoundf("%q not in whitelist", host)
	}
	return nil
}

func (s *server) autoTLSConfig() *tls.Config {
	autoCert := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: s.hostPolicy,
		Cache:      s.autoCertCache,
	}
	tlsConfig := autoCert.TLSConfig()
	delayer := &certdelay.Delayer{
		CertForDomain: func(domain string) *tls.Certificate {
			return s.manualCerts[domain]
		},
	}
	tlsConfig.GetCertificate = delayer.Wrap(tlsConfig.GetCertificate)

	return tlsConfig
}
