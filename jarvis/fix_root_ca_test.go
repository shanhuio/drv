package jarvis

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"shanhu.io/drv/burmilla"
	"shanhu.io/std/docker"
	"shanhu.io/std/docker/dockertest"
)

// newFileConsole starts a fake docker daemon with a "console" container whose
// filesystem is seeded with files and whose execs are emulated by execFn
// (nil for the daemon's default exit-0 response), and returns a Burmilla stub
// wired to it.
func newFileConsole(
	t *testing.T,
	files map[string][]byte,
	execFn func(cmd []string) dockertest.ExecResponse,
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
	d.AddContainer(&dockertest.Container{
		ID: "console", Files: files, ExecFunc: execFn,
	})
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

	var cpCmd []string
	b := newFileConsole(
		t,
		map[string][]byte{rootCACertFile: broken},
		func(cmd []string) dockertest.ExecResponse {
			if len(cmd) > 0 && cmd[0] == "sudo" {
				cpCmd = cmd
			}
			return dockertest.ExecResponse{}
		},
	)
	fixed, err := fixRootCACertificates(b)
	if err != nil {
		t.Fatalf("fixRootCACertificates: %v", err)
	}
	if !fixed {
		t.Errorf("fixRootCACertificates = false, want true on broken bundle")
	}

	// The new bundle is staged in /tmp before being copied into place.
	got := readConsoleFile(t, b, "/tmp/ca-certificates-202606.crt")
	if !bytes.Equal(got, caCertificates202606) {
		t.Errorf(
			"staged file mismatch: got %d bytes, want embedded %d bytes",
			len(got), len(caCertificates202606),
		)
	}

	// The destination is a mounted file, so it is overwritten with sudo cp.
	wantCmd := []string{
		"sudo", "cp",
		"/tmp/ca-certificates-202606.crt", rootCACertFile,
	}
	if !reflect.DeepEqual(cpCmd, wantCmd) {
		t.Errorf("overwrite cmd = %q, want %q", cpCmd, wantCmd)
	}
}

func TestFixRootCACertificatesLeavesOthers(t *testing.T) {
	const content = "some other ca bundle\n"
	b := newFileConsole(t, map[string][]byte{
		rootCACertFile: []byte(content),
	}, nil)

	fixed, err := fixRootCACertificates(b)
	if err != nil {
		t.Fatalf("fixRootCACertificates: %v", err)
	}
	if fixed {
		t.Errorf("fixRootCACertificates = true, want false on other bundle")
	}

	got := readConsoleFile(t, b, rootCACertFile)
	if string(got) != content {
		t.Errorf("file changed: got %q, want %q", got, content)
	}
}

func TestFixRootCACertificatesReadError(t *testing.T) {
	// The console has no such file.
	b := newFileConsole(t, nil, nil)
	if _, err := fixRootCACertificates(b); err == nil {
		t.Fatalf("fixRootCACertificates: got nil error, want read error")
	}
}

func TestFixRootCACertificatesOverwriteError(t *testing.T) {
	broken, err := os.ReadFile(
		filepath.Join("testdata", "ca-certificates-2025.crt.rancher"),
	)
	if err != nil {
		t.Fatalf("read broken bundle: %v", err)
	}

	// The sudo cp fails.
	b := newFileConsole(
		t,
		map[string][]byte{rootCACertFile: broken},
		func(cmd []string) dockertest.ExecResponse {
			if len(cmd) > 0 && cmd[0] == "sudo" {
				return dockertest.ExecResponse{ExitCode: 1}
			}
			return dockertest.ExecResponse{}
		},
	)
	if _, err := fixRootCACertificates(b); err == nil {
		t.Fatalf("fixRootCACertificates: got nil error, want overwrite error")
	}
}
