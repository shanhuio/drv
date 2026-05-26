package burmilla

import (
	"strings"

	"shanhu.io/std/errcode"
)

// HostIPs returns the ip addresses of the given network device.
func HostIPs(b *Burmilla, dev string) ([]string, error) {
	args := strings.Fields(
		"ip -br -family inet address show dev",
	)
	args = append(args, dev)
	out, err := b.ExecOutput(args)
	if err != nil {
		return nil, err
	}

	// Output is in form of:
	//   eth0    UP    x.x.x.x/xx x.x.x.x/xx ...
	// fields[2:] should all be IPv4 addresses
	s := string(out)
	fields := strings.Fields(s)
	if len(fields) < 3 {
		return nil, errcode.Internalf("unexpected output: %q", s)
	}

	ips := append([]string{}, fields[2:]...)
	return ips, nil
}
