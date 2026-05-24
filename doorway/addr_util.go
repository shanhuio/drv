package doorway

import (
	"net"
)

func lisAddr(lis net.Listener) string {
	return lis.Addr().String()
}
