package jarvis

import (
	"reflect"
	"testing"

	"shanhu.io/drv/burmilla"
	"shanhu.io/std/docker/dockertest"
)

// newConsoleBurmilla starts a fake docker daemon with a "console" container
// whose execs are emulated by fn, and returns a Burmilla stub wired to it.
func newConsoleBurmilla(
	t *testing.T, fn func(cmd []string) dockertest.ExecResponse,
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
	d.AddContainer(&dockertest.Container{ID: "console", ExecFunc: fn})
	return burmilla.New(d.Client)
}

const osUpdateURL = "https://www.homedrive.io/os.yml"

func TestSetOSUpdateSourceAlreadySet(t *testing.T) {
	var setCmds [][]string
	b := newConsoleBurmilla(t, func(cmd []string) dockertest.ExecResponse {
		if len(cmd) >= 3 && cmd[1] == "config" && cmd[2] == "get" {
			return dockertest.ExecResponse{Stdout: osUpdateURL + "\n"}
		}
		if len(cmd) >= 3 && cmd[1] == "config" && cmd[2] == "set" {
			setCmds = append(setCmds, cmd)
		}
		return dockertest.ExecResponse{}
	})

	if err := setOSUpdateSource(b); err != nil {
		t.Fatalf("setOSUpdateSource: %v", err)
	}
	// Source already matches, so no config set should be issued.
	if len(setCmds) != 0 {
		t.Errorf("issued config set %q, want none", setCmds)
	}
}

func TestSetOSUpdateSourceUpdates(t *testing.T) {
	var setCmds [][]string
	b := newConsoleBurmilla(t, func(cmd []string) dockertest.ExecResponse {
		if len(cmd) >= 3 && cmd[1] == "config" && cmd[2] == "get" {
			return dockertest.ExecResponse{Stdout: "https://old.example.com/os.yml\n"}
		}
		if len(cmd) >= 3 && cmd[1] == "config" && cmd[2] == "set" {
			setCmds = append(setCmds, cmd)
		}
		return dockertest.ExecResponse{}
	})

	if err := setOSUpdateSource(b); err != nil {
		t.Fatalf("setOSUpdateSource: %v", err)
	}

	want := [][]string{{
		"ros", "config", "set", "rancher.upgrade.url", osUpdateURL,
	}}
	if !reflect.DeepEqual(setCmds, want) {
		t.Errorf("config set commands = %q, want %q", setCmds, want)
	}
}

func TestSetOSUpdateSourceGetError(t *testing.T) {
	b := newConsoleBurmilla(t, func(cmd []string) dockertest.ExecResponse {
		return dockertest.ExecResponse{ExitCode: 1}
	})

	if err := setOSUpdateSource(b); err == nil {
		t.Fatalf("setOSUpdateSource: got nil error, want error reading config")
	}
}

func TestSetOSUpdateSourceSetError(t *testing.T) {
	b := newConsoleBurmilla(t, func(cmd []string) dockertest.ExecResponse {
		if len(cmd) >= 3 && cmd[1] == "config" && cmd[2] == "get" {
			return dockertest.ExecResponse{Stdout: "https://old.example.com/os.yml\n"}
		}
		// The config set fails.
		return dockertest.ExecResponse{ExitCode: 1}
	})

	if err := setOSUpdateSource(b); err == nil {
		t.Fatalf("setOSUpdateSource: got nil error, want error setting config")
	}
}

func TestIsUEFI(t *testing.T) {
	for _, test := range []struct {
		exitCode int
		want     bool
	}{
		{exitCode: 0, want: true},  // /sys/firmware/efi exists
		{exitCode: 1, want: false}, // it does not
	} {
		var gotCmd []string
		b := newConsoleBurmilla(t, func(cmd []string) dockertest.ExecResponse {
			gotCmd = cmd
			return dockertest.ExecResponse{ExitCode: test.exitCode}
		})

		got, err := isUEFI(b)
		if err != nil {
			t.Fatalf("isUEFI (exit %d): %v", test.exitCode, err)
		}
		if want := []string{
			"test", "-d", "/sys/firmware/efi",
		}; !reflect.DeepEqual(gotCmd, want) {
			t.Errorf("exec cmd = %q, want %q", gotCmd, want)
		}
		if got != test.want {
			t.Errorf("isUEFI (exit %d) = %v, want %v", test.exitCode, got, test.want)
		}
	}
}
