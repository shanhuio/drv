package sniproxy

import (
	"context"
	"net/url"

	"github.com/gorilla/websocket"
	"shanhu.io/std/errcode"
)

// DialOption provides addition option for dialing.
type DialOption struct {
	// Path is the path of the WebSocket endpoint.
	Path string

	// Dialer is an optional WebSocket dialer to use.
	Dialer *websocket.Dialer

	// TunnelOptions fine tunes the behavior of a tunnel.
	TunnelOptions *Options

	// WithoutTLS uses the "ws://" scheme rather than the "wss://" scheme.
	WithoutTLS bool
}

// Dial connects to fabrics server, establishes a tunnel and returns an
// endpoint.
func Dial(
	ctx context.Context, r Router, opt *DialOption,
) (*Endpoint, error) {
	if opt == nil {
		opt = &DialOption{}
	}

	host, token, err := r.Route(ctx)
	if err != nil {
		return nil, errcode.Annotate(err, "proxy route")
	}

	addr := &url.URL{Scheme: "wss", Host: host, Path: opt.Path}
	if opt.WithoutTLS {
		addr.Scheme = "ws"
	}
	dialer := &websocketDialer{
		url:    addr,
		token:  token,
		dialer: opt.Dialer,
	}
	tunnOpt := opt.TunnelOptions
	if tunnOpt == nil {
		tunnOpt = &Options{}
	}
	conn, err := dialer.dial(ctx, tunnOpt)
	if err != nil {
		return nil, err
	}
	return newEndpoint(conn, dialer, tunnOpt), nil
}
