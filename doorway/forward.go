package doorway

import (
	"net/http"
	"net/url"
)

func forwardToHTTP(req *http.Request, host string) {
	req.Header.Set("X-Forwarded-Host", req.Host)

	// The remote address that doorway sees is always the real address, as it
	// is either directly listening on the https port, or it gets the remote
	// address from the fabrics tunnel. This avoids IP spoofing that might
	// confuse Nextcloud or other internal applications.
	req.Header.Del("X-Forwarded-For")
	req.Header.Del("X-Real-IP")

	if len(req.RemoteAddr) > 0 && req.RemoteAddr[0] == '|' {
		req.RemoteAddr = req.RemoteAddr[1:] // trim the '|'
		req.Header.Add("Via", "1.0 hometunn")
	}

	u := req.URL      // Modify the URL.
	u.Scheme = "http" // Terminates http.
	u.Host = host
}

var sinkURL = &url.URL{
	Scheme: "http",
	Host:   "localhost",
}
