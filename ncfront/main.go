// Package ncfront provides a reverse proxy that runs in front of nextcloud to
// handle several nextcloud docker issues.
package ncfront

import (
	"flag"
	"log"
	"net/http"
	"os"
)

// Main is the main entrance of the nextcloud front proxy.
func Main() {
	destHost := flag.String("nextcloud", "", "nextcloud's host")
	addr := flag.String("addr", ":8080", "address to listen on")
	flag.Parse()

	if *destHost == "" {
		if v, ok := os.LookupEnv("NEXTCLOUD"); ok {
			*destHost = v
		} else {
			const defaultAddr = "nextcloud:80"
			*destHost = defaultAddr
		}
	}

	f := newFront(*destHost)
	log.Printf("listening on %s, forwarding to %s", *addr, *destHost)
	if err := http.ListenAndServe(*addr, f); err != nil {
		log.Fatal(err)
	}
}
