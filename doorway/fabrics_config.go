package doorway

import (
	"log"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	fabdial "shanhu.io/drv/fabricsdial"
	"shanhu.io/drv/homedial"
	"shanhu.io/g/https/httpstest"
	"shanhu.io/g/sniproxy"
	"shanhu.io/std/errcode"
)

// FabricsConfig has the configuration for connecting HomeDrive Fabrics.
// This config is JSON marshallable.
type FabricsConfig struct {
	User string
	Host string `json:",omitempty"` // Default using fabrics.homedrive.io

	InsecurelyDialTo string `json:",omitempty"`
}

func (c *FabricsConfig) host() string {
	if c.Host == "" {
		return "fabrics.homedrive.io"
	}
	return c.Host
}

type fabricsConfig struct {
	// Explicit dialer creater. Will use this dialer instead of the User:Host
	// when this is explicitly specified.
	dialer *fabdial.Dialer

	*FabricsConfig
	identity Identity
}

func makeFabricsDialer(ctx C, config *fabricsConfig) (
	*fabdial.Dialer, error,
) {
	if config.dialer != nil {
		return config.dialer, nil
	}

	key, err := config.identity.Load(ctx)
	if err != nil {
		return nil, errcode.Annotate(err, "read fabrics key")
	}

	router := &fabdial.SimpleRouter{
		Host: config.host(),
		User: config.User,
		Key:  key,
	}
	dialer := &fabdial.Dialer{Router: router}

	if config.InsecurelyDialTo != "" {
		tr := httpstest.InsecureSink(config.InsecurelyDialTo)
		router.Transport = tr
		dialer.WebSocketDialer = fabdial.NewWebSocketDialer(tr)
	} else {
		router.Transport = &http.Transport{DialContext: homedial.Dial}
		dialer.WebSocketDialer = &websocket.Dialer{
			NetDialContext:  homedial.Dial,
			ReadBufferSize:  sniproxy.DefaultReadBufferSize,
			WriteBufferSize: sniproxy.DefaultWriteBufferSize,
		}
	}
	return dialer, nil
}

func listenFabrics(ctx C, config *fabricsConfig) (*tagListener, error) {
	d, err := makeFabricsDialer(ctx, config)
	if err != nil {
		return nil, err
	}
	lis, err := newReconnectListener(
		func() (net.Listener, error) {
			ep, err := d.Dial(ctx)
			if err != nil {
				return nil, errcode.Annotatef(err, "dial proxy")
			}
			return &fabricsListener{Endpoint: ep}, nil
		},
		func(err error) { log.Println("fabrics connection: ", err) },
	)
	if err != nil {
		return nil, errcode.Annotatef(err, "dial fabrics")
	}
	return newTagListener(lis, tagFabrics), nil
}
