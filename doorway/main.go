// Package doorway is the HTTP frontend on a shanhu instance.
package doorway

import (
	"context"
	"flag"
	"log"
)

// Main is the main entrance for doorway binary.
func Main() {
	var (
		httpsAddr = flag.String("https", ":8443", "HTTPS address to listen on.")
		httpAddr  = flag.String("http", ":8080", "HTTP address to listen on.")
		home      = flag.String("home", ".", "home directory")
	)
	flag.Parse()

	ctx := context.Background()

	config, err := ConfigFromHome(*home)
	if err != nil {
		log.Fatal(err)
	}

	config.LocalAddr = *httpsAddr
	if *httpAddr != "" {
		config.HTTPServer.Addr = *httpAddr
	}

	if err := Serve(ctx, config); err != nil {
		log.Fatal(err)
	}
}
