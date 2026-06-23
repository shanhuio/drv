package jarvis

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"shanhu.io/drv/burmilla"
	"shanhu.io/std/docker"
	"shanhu.io/std/docker/dockertest"
)

// newFileConsole starts a fake docker daemon with a "console" container whose
// filesystem is seeded with files, and returns a Burmilla stub wired to it.
func newFileConsole(
	t *testing.T, files map[string][]byte,
) *burmilla.Burmilla {
	t.Helper()
	d, err := dockertest.New()
	if err != nil {
		t.Fatalf("dockertest.New: %v", err)
	}
	t.Cleanup(func() {
		if err := d.Close(); err != nil {
			t.Errorf("close fake daemon: %v", err)
		}
		if err := d.Err(); err != nil {
			t.Errorf("fake daemon: %v", err)
		}
	})
	d.AddContainer(&dockertest.Container{ID: "console", Files: files})
	return burmilla.New(d.Client)
}

// readConsoleFile reads a file back out of the fake console container.
func readConsoleFile(t *testing.T, b *burmilla.Burmilla, path string) []byte {
	t.Helper()
	bs, err := docker.ReadContFile(b.Console(), path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	return bs
}

func TestFixRootCACertificatesReplaces(t *testing.T) {
	broken, err := os.ReadFile(
		filepath.Join("testdata", "ca-certificates-2025.crt.rancher"),
	)
	if err != nil {
		t.Fatalf("read broken bundle: %v", err)
	}

	b := newFileConsole(t, map[string][]byte{rootCACertFile: broken})
	if err := fixRootCACertificates(b); err != nil {
		t.Fatalf("fixRootCACertificates: %v", err)
	}

	got := readConsoleFile(t, b, rootCACertFile)
	if !bytes.Equal(got, caCertificates202606) {
		t.Errorf(
			"file not replaced: got %d bytes, want embedded %d bytes",
			len(got), len(caCertificates202606),
		)
	}
}

func TestFixRootCACertificatesLeavesOthers(t *testing.T) {
	const content = "some other ca bundle\n"
	b := newFileConsole(t, map[string][]byte{
		rootCACertFile: []byte(content),
	})

	if err := fixRootCACertificates(b); err != nil {
		t.Fatalf("fixRootCACertificates: %v", err)
	}

	got := readConsoleFile(t, b, rootCACertFile)
	if string(got) != content {
		t.Errorf("file changed: got %q, want %q", got, content)
	}
}

func TestFixRootCACertificatesReadError(t *testing.T) {
	// The console has no such file.
	b := newFileConsole(t, nil)
	if err := fixRootCACertificates(b); err == nil {
		t.Fatalf("fixRootCACertificates: got nil error, want read error")
	}
}
