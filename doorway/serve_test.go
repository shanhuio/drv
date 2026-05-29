package doorway

import (
	"io"
	"testing"

	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"

	"shanhu.io/g/https/httpstest"
	"shanhu.io/std/jsonx"
)

func checkGet(t *testing.T, c *http.Client, url, want string) {
	resp, err := c.Get(url)
	if err != nil {
		t.Errorf("get %s: %s", url, err)
		return
	}
	defer resp.Body.Close()

	bs, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("read body: %s", err)
		return
	}

	if string(bs) != want {
		t.Errorf("get %s, want %q, got %q", url, want, string(bs))
	}
}

func TestServe(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		t.Logf("request: %q", req.URL)
		for k := range req.Header {
			v := req.Header.Get(k)
			t.Logf("%q=%q", k, v)
		}
		fmt.Fprint(w, "dest")
	})
	s := httptest.NewServer(h)

	hostMap := map[string]string{
		"ctrl.shanhu.io": HomeHost,
		"shanhu.io":      lisAddr(s.Listener),
	}

	doorwayHome, err := os.MkdirTemp("", "doorway")
	if err != nil {
		t.Fatal("make doorway temp home:", err)
	}
	defer os.RemoveAll(doorwayHome)

	doorwayEtc := filepath.Join(doorwayHome, "etc")
	if err := os.MkdirAll(doorwayEtc, 0700); err != nil {
		t.Fatal("make doorway etc:", err)
	}
	doorwayVar := filepath.Join(doorwayHome, "var")
	if err := os.MkdirAll(doorwayVar, 0700); err != nil {
		t.Fatal("make doorway var:", err)
	}

	mapFile := filepath.Join(doorwayEtc, "host-map.jsonx")
	if err := jsonx.WriteFile(mapFile, hostMap); err != nil {
		t.Fatal("write host map", err)
	}

	tlsConfigs, err := httpstest.NewTLSConfigs(
		[]string{"ctrl.shanhu.io", "shanhu.io"},
	)
	if err != nil {
		t.Fatal(err)
	}
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	defer lis.Close()

	config, err := ConfigFromHome(doorwayHome)
	if err != nil {
		t.Fatal("read config:", err)
	}
	internal := makeInternalConfig(config)
	internal.listen = &listenConfig{
		local: &localListenConfig{listener: lis},
	}
	internal.tlsConfig = tlsConfigs.Server

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bgErr := make(chan error, 1)
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		bgErr <- serve(ctx, internal)
	}(ctx)

	client := &http.Client{
		Transport: tlsConfigs.Sink(lisAddr(lis)),
	}
	checkGet(t, client, "https://shanhu.io", "dest")
	checkGet(t, client, "https://shanhu.io/subpage", "dest")
	checkGet(t, client, "https://ctrl.shanhu.io/health", "ok")

	cancel()
	if err := <-bgErr; err != nil {
		if err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}
}
