package fabricsdial

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"shanhu.io/g/sniproxy"
)

// NewWebSocketDialer creates a new WebSocket dialer from
// a http transport.
func NewWebSocketDialer(tr *http.Transport) *websocket.Dialer {
	return &websocket.Dialer{
		NetDialContext:  tr.DialContext,
		TLSClientConfig: tr.TLSClientConfig,
	}
}

// Dialer dials to a HomeDrive Fabrics service.
type Dialer struct {
	Router          sniproxy.Router
	WebSocketDialer *websocket.Dialer
	TunnelOptions   *sniproxy.Options
}

var defaultTunnelOptions = &sniproxy.Options{
	Siding:       true,
	DialWithAddr: true,
}

// Dial connects to a HomeDrive Fabrics service, and returns
// an SNI-proxy endpoint.
func (d *Dialer) Dial(ctx context.Context) (*sniproxy.Endpoint, error) {
	tunnOpts := d.TunnelOptions
	if tunnOpts == nil {
		tunnOpts = defaultTunnelOptions
	}
	opt := &sniproxy.DialOption{
		Path:          "/endpoint",
		TunnelOptions: tunnOpts,
		Dialer:        d.WebSocketDialer,
	}
	return sniproxy.Dial(ctx, d.Router, opt)
}
