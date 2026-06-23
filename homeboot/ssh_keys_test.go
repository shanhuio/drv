package homeboot

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"shanhu.io/drv/drvapi"
	"shanhu.io/g/httputil"
	"shanhu.io/std/fakeweb"
)

// fakeClient starts an HTTPS fakeweb server serving handler at domain and
// returns an httputil.Client wired to talk to it (trusting the server's
// self-signed cert). The server is closed at the end of the test.
func fakeClient(
	t *testing.T, domain string, handler http.Handler,
) *httputil.Client {
	t.Helper()
	s, err := fakeweb.NewServer(domain, handler)
	if err != nil {
		t.Fatalf("fakeweb.NewServer(%q): %v", domain, err)
	}
	t.Cleanup(func() {
		if err := s.Close(); err != nil {
			t.Errorf("close fakeweb server: %v", err)
		}
	})
	return &httputil.Client{
		Server:    &url.URL{Scheme: "https", Host: domain},
		Transport: s.Client().Transport,
	}
}

func TestFetchGitHubKeys(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/alice.keys" {
			http.NotFound(w, r)
			return
		}
		// Surrounding whitespace and blank lines must be trimmed/dropped.
		io.WriteString(w, "ssh-rsa key1\n\n  ssh-rsa key2  \n")
	})

	c := fakeClient(t, "github.com", handler)
	got, err := fetchGitHubKeys(c, "alice")
	if err != nil {
		t.Fatalf("fetchGitHubKeys: %v", err)
	}
	want := []string{"ssh-rsa key1", "ssh-rsa key2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("fetchGitHubKeys = %q, want %q", got, want)
	}
}

func TestFetchGitHubKeysNotFound(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	c := fakeClient(t, "github.com", handler)
	if _, err := fetchGitHubKeys(c, "ghost"); err == nil {
		t.Fatalf("fetchGitHubKeys: got nil error, want error on 404")
	}
}

func TestFetchUserKeys(t *testing.T) {
	var gotUser string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pubapi/user/sshkeys" {
			http.NotFound(w, r)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotUser); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.NewEncoder(w).Encode(&drvapi.UserSSHKeyLines{
			Keys: []string{"ssh-rsa k1", "ssh-rsa k2"},
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	})

	c := fakeClient(t, "www.homedrive.io", handler)
	got, err := fetchUserKeys(c, "bob")
	if err != nil {
		t.Fatalf("fetchUserKeys: %v", err)
	}
	if gotUser != "bob" {
		t.Errorf("server received user %q, want %q", gotUser, "bob")
	}
	want := []string{"ssh-rsa k1", "ssh-rsa k2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("fetchUserKeys = %q, want %q", got, want)
	}
}

func TestFetchUserKeysError(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})

	c := fakeClient(t, "www.homedrive.io", handler)
	if _, err := fetchUserKeys(c, "bob"); err == nil {
		t.Fatalf("fetchUserKeys: got nil error, want error on 500")
	}
}
