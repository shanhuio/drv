package sniproxy

import (
	"testing"

	"context"
	"net"
	"net/http"
	"net/http/httptest"

	"shanhu.io/g/aries"
	"shanhu.io/g/https/httpstest"
	"shanhu.io/g/httputil"
)

func TestProxy(t *testing.T) {
	tlsConfigs, err := httpstest.NewTLSConfigs([]string{
		"site1.com",
		"site2.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	site1 := httptest.NewUnstartedServer(aries.StringFunc("site1"))
	site1.TLS = tlsConfigs.Server
	site1.StartTLS()
	defer site1.Close()

	site2 := httptest.NewUnstartedServer(aries.StringFunc("site2"))
	site2.TLS = tlsConfigs.Server
	site2.StartTLS()
	defer site2.Close()

	httpAddr := func(s *httptest.Server) *net.TCPAddr {
		return s.Listener.Addr().(*net.TCPAddr)
	}

	dialer := &tcpDialer{
		raddrs: map[string]*net.TCPAddr{
			"site1.com": httpAddr(site1),
			"site2.com": httpAddr(site2),
		},
	}

	localhost := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)}
	lis, err := net.ListenTCP("tcp", localhost)
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	p := newProxy(dialer)
	ctx := context.Background()
	go p.serve(ctx, lis)

	client := &http.Client{
		Transport: tlsConfigs.Sink(lis.Addr().String()),
	}

	got, err := httputil.GetString(client, "https://site1.com")
	if err != nil {
		t.Fatal("get site1:", err)
	}
	if got != "site1" {
		t.Errorf("want site1, got %q", got)
	}

	got, err = httputil.GetString(client, "https://site2.com")
	if err != nil {
		t.Fatal("get site2:", err)
	}
	if got != "site2" {
		t.Errorf("want site2, got %q", got)
	}
}
