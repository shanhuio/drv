package sniproxy

import (
	"context"
)

// Router provides a host to connect with a token.
type Router interface {
	Route(ctx context.Context) (host string, token string, err error)
}

// StaticRouter routes to the given host with the given token.
type StaticRouter struct {
	Host  string
	Token string
}

// Route returns the given static host and token.
func (r *StaticRouter) Route(ctx context.Context) (string, string, error) {
	return r.Host, r.Token, nil
}
