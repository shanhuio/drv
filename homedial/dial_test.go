package homedial

import (
	"testing"
)

func TestMapAddress(t *testing.T) {
	for _, test := range []struct {
		net, addr, addrWant string
	}{{
		net:      "tcp",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "178.128.130.77:443",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "178.128.130.77:443",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io:80",
		addrWant: "178.128.130.77:80",
	}, {
		net:      "tcp4",
		addr:     "fabrics.homedrive.io.:80",
		addrWant: "178.128.130.77:80",
	}, {
		net:      "tcp6",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "fabrics.homedrive.io:443",
	}, {
		net:      "udp",
		addr:     "fabrics.homedrive.io:443",
		addrWant: "fabrics.homedrive.io:443",
	}, {
		// Other mapped hosts resolve too.
		net:      "tcp",
		addr:     "www.homedrive.io:443",
		addrWant: "167.172.10.171:443",
	}, {
		net:      "tcp4",
		addr:     "fabrics-sgp.homedrive.io:22",
		addrWant: "149.28.152.149:22",
	}, {
		// Hosts not in the map are left unchanged.
		net:      "tcp",
		addr:     "unknown.example.com:443",
		addrWant: "unknown.example.com:443",
	}, {
		// A trailing dot on an unmapped host is not stripped from the result.
		net:      "tcp",
		addr:     "unknown.example.com.:443",
		addrWant: "unknown.example.com.:443",
	}, {
		// IP literals are passed through, even when on the right network.
		net:      "tcp",
		addr:     "1.2.3.4:443",
		addrWant: "1.2.3.4:443",
	}, {
		// Addresses without a port fail to split and are returned as-is.
		net:      "tcp",
		addr:     "fabrics.homedrive.io",
		addrWant: "fabrics.homedrive.io",
	}, {
		net:      "tcp",
		addr:     "",
		addrWant: "",
	}} {
		got := mapAddress(test.net, test.addr)
		if got != test.addrWant {
			t.Errorf(
				"map net=%s addr=%q, got %q, want %q",
				test.net, test.addr, test.addrWant, got,
			)
		}
	}
}
