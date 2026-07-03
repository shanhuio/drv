package ncfront

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type front struct {
	proxy *httputil.ReverseProxy
}

func newFront(destHost string) *front {
	u := &url.URL{Scheme: "http", Host: destHost}
	p := httputil.NewSingleHostReverseProxy(u)
	return &front{proxy: p}
}

var redirects = map[string]string{
	"/.well-known/caldav":  "/remote.php/dav/",
	"/.well-known/carddav": "/remote.php/dav/",
}

func (f *front) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	re, found := redirects[r.URL.Path]
	if found {
		http.Redirect(w, r, re, http.StatusMovedPermanently)
		return
	}
	f.proxy.ServeHTTP(w, r)
}
