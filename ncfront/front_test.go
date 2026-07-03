package ncfront

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFrontRedirects(t *testing.T) {
	f := newFront("nextcloud:80")

	for path, want := range redirects {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		w := httptest.NewRecorder()
		f.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMovedPermanently {
			t.Errorf(
				"GET %s: status = %d, want %d",
				path, resp.StatusCode, http.StatusMovedPermanently,
			)
		}
		if got := resp.Header.Get("Location"); got != want {
			t.Errorf("GET %s: Location = %q, want %q", path, got, want)
		}
	}
}

func TestFrontProxies(t *testing.T) {
	const body = "hello from nextcloud"
	var gotPath string
	backend := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			io.WriteString(w, body)
		},
	))
	defer backend.Close()

	// The backend URL is http://host:port; strip the scheme for destHost.
	destHost := backend.Listener.Addr().String()
	f := newFront(destHost)

	req := httptest.NewRequest(http.MethodGet, "/index.php", nil)
	w := httptest.NewRecorder()
	f.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if gotPath != "/index.php" {
		t.Errorf("backend saw path %q, want %q", gotPath, "/index.php")
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(got) != body {
		t.Errorf("body = %q, want %q", got, body)
	}
}

func TestFrontProxiesWellKnownNonRedirect(t *testing.T) {
	// A /.well-known path that is not in the redirect table is proxied
	// through rather than redirected.
	var gotPath string
	backend := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
		},
	))
	defer backend.Close()

	f := newFront(backend.Listener.Addr().String())

	const path = "/.well-known/webfinger"
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	f.ServeHTTP(w, req)

	if resp := w.Result(); resp.StatusCode == http.StatusMovedPermanently {
		t.Errorf("GET %s: got redirect, want proxy pass-through", path)
	}
	if gotPath != path {
		t.Errorf("backend saw path %q, want %q", gotPath, path)
	}
}
