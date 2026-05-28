package doorway

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"

	fabdial "shanhu.io/drv/fabricsdial"
	"shanhu.io/drv/sniproxy"
	"shanhu.io/g/aries"
	"shanhu.io/g/netutil"
	"shanhu.io/std/errcode"
)

// Config is the config of a doorway.
type Config struct {
	// Server is the config for the http server.
	// This also includes the reverse proxy.
	Server *ServerConfig

	// TLSProxy is the configuration for the TLS proxy.
	TLSProxy *TLSProxyConfig

	// HTTPServer is the config for the http server part.
	HTTPServer *HTTPServerConfig

	// Local address to listen on.
	LocalAddr string

	Fabrics         *FabricsConfig // Config for dialing fabrics.
	FabricsIdentity Identity       // Identity for dialing fabrics.

	// Alternative fabrics dialer.
	FabricsDialer *fabdial.Dialer

	// TLSConfig is for the TLS config for serving the service via https.
	// If not specified, autocert from Letsencrypt will be used.
	TLSConfig *tls.Config

	// ListenDone is the callback function when listen is done.
	ListenDone func()
}

type internalConfig struct {
	server     *ServerConfig
	tlsProxy   *TLSProxyConfig
	listen     *listenConfig
	listenDone func()
	tlsConfig  *tls.Config
}

func makeInternalConfig(config *Config) *internalConfig {
	lisConfig := new(listenConfig)
	if config.LocalAddr != "" {
		lisConfig.local = &localListenConfig{addr: config.LocalAddr}
	}
	if config.FabricsDialer != nil {
		lisConfig.fabrics = &fabricsConfig{
			dialer: config.FabricsDialer,
		}
	} else if config.Fabrics != nil {
		lisConfig.fabrics = &fabricsConfig{
			FabricsConfig: config.Fabrics,
			identity:      config.FabricsIdentity,
		}
	}

	return &internalConfig{
		server:     config.Server,
		tlsProxy:   config.TLSProxy,
		listen:     lisConfig,
		tlsConfig:  config.TLSConfig,
		listenDone: config.ListenDone,
	}
}

// Serve serves doorway with the given config.
func Serve(ctx C, config *Config) error {
	if config.HTTPServer != nil {
		http := newHTTPServer(config.HTTPServer)
		go runHTTPServer(http)
	}

	internal := makeInternalConfig(config)
	return serve(ctx, internal)
}

func serve(ctx C, config *internalConfig) error {
	server, err := newServer(config.server)
	if err != nil {
		return errcode.Annotate(err, "make server")
	}

	lis, err := listen(ctx, config.listen)
	if err != nil {
		return errcode.Annotate(err, "listen")
	}

	var httpsLis net.Listener = lis
	if config.tlsProxy != nil {
		httpsLis = newTLSProxy(lis, config.tlsProxy)
	}
	defer httpsLis.Close()

	if config.listenDone != nil {
		config.listenDone()
	}

	tlsConfig := config.tlsConfig
	if tlsConfig == nil {
		tlsConfig = server.autoTLSConfig()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log.Printf("starts https on %q", lisAddr(httpsLis))
	https := &http.Server{
		TLSConfig: tlsConfig,
		Handler:   aries.Serve(server),
	}
	go func() {
		<-ctx.Done()
		https.Close()
	}()

	keepAlive := netutil.WrapKeepAlive(httpsLis)
	if err := https.ServeTLS(keepAlive, "", ""); err != nil {
		if sniproxy.IsClosedConnError(err) {
			return nil
		}
		return err
	}
	return nil
}
