package sniproxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
	"shanhu.io/g/errcode"
	"shanhu.io/g/httputil"
)

type websocketDialer struct {
	url    *url.URL
	token  string
	dialer *websocket.Dialer
}

func (d *websocketDialer) address() string {
	return d.url.String()
}

func (d *websocketDialer) getDialer() *websocket.Dialer {
	if d.dialer != nil {
		return d.dialer
	}
	return &websocket.Dialer{
		ReadBufferSize:  DefaultReadBufferSize,
		WriteBufferSize: DefaultWriteBufferSize,
	}
}

func (d *websocketDialer) dial(ctx context.Context, opt *Options) (
	*websocket.Conn, error,
) {
	header := make(http.Header)
	if d.token != "" {
		httputil.SetAuthToken(header, d.token)
	}

	u := *d.url
	q := u.Query()
	if opt == nil {
		opt = new(Options)
	}
	optBytes, err := json.Marshal(opt)
	if err != nil {
		return nil, errcode.Annotate(err, "marshal options")
	}
	q.Set("opt", string(optBytes))

	u.RawQuery = q.Encode()

	dialer := d.getDialer()
	conn, _, err := dialer.DialContext(ctx, u.String(), header)
	return conn, err
}

func (d *websocketDialer) dialSide(
	ctx context.Context, tok string, k *sessionKey,
) (*websocket.Conn, error) {
	side, err := k.encode()
	if err != nil {
		return nil, errcode.Annotate(err, "encode session key")
	}

	header := make(http.Header)
	if tok != "" {
		httputil.SetAuthToken(header, tok)
	}

	u := *d.url
	q := u.Query()
	q.Set("side", side)
	u.RawQuery = q.Encode()

	dialer := d.getDialer()
	conn, _, err := dialer.DialContext(ctx, u.String(), header)
	return conn, err
}
