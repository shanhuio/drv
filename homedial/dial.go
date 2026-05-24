package homedial

import (
	"context"
	"net"
	"strings"
	"time"
)

var fallbackNetDialer = &net.Dialer{
	Timeout:   10 * time.Second,
	KeepAlive: 30 * time.Second,
}

var homedrvIPv4 = map[string]string{
	"homedrive.io":             "167.172.10.171",
	"www.homedrive.io":         "167.172.10.171",
	"fabrics.homedrive.io":     "178.128.130.77",
	"fabrics-ge.homedrive.io":  "157.245.24.167",
	"fabrics-ge1.homedrive.io": "206.81.25.26",
	"fabrics-sgp.homedrive.io": "149.28.152.149",
}

func mapAddress(network, addr string) string {
	if !(network == "tcp" || network == "tcp4") {
		return addr
	}
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	// Manually resolve IPv4 addresses for fabrics. This by passes DNS
	// resolvers in user's home networks, which might be faulty.
	trimmed := strings.TrimSuffix(host, ".")
	if ip, ok := homedrvIPv4[trimmed]; ok {
		// Directly resolve to IP address.
		return net.JoinHostPort(ip, port)
	}
	return addr
}

// Dial dials HomeDrive servers.
func Dial(ctx context.Context, network, addr string) (
	net.Conn, error,
) {
	addr = mapAddress(network, addr)
	return fallbackNetDialer.DialContext(ctx, network, addr)
}
