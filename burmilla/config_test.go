package burmilla

import (
	"reflect"
	"strings"
	"testing"

	"shanhu.io/drv/bosinit"
	"shanhu.io/std/docker/dockertest"
)

// newExecDaemon starts a fake docker daemon with a "console" container whose
// execs are emulated by fn. It registers cleanup that closes the daemon and
// surfaces any internal errors it recorded.
func newExecDaemon(
	t *testing.T, fn func(cmd []string) dockertest.ExecResponse,
) *dockertest.FakeDaemon {
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
	return d
}

func TestConfigExport(t *testing.T) {
	const out = "ssh_authorized_keys:\n- ssh-rsa key1\n- ssh-rsa key2\n"

	var gotCmd []string
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		gotCmd = cmd
		return dockertest.ExecResponse{Stdout: out}
	})

	b := New(d.Client)
	config, err := ConfigExport(b)
	if err != nil {
		t.Fatalf("ConfigExport: %v", err)
	}

	if want := []string{"ros", "config", "export"}; !reflect.DeepEqual(
		gotCmd, want,
	) {
		t.Errorf("exec cmd = %q, want %q", gotCmd, want)
	}
	want := []string{"ssh-rsa key1", "ssh-rsa key2"}
	if got := config.SSHAuthorizedKeys; !reflect.DeepEqual(got, want) {
		t.Errorf("SSHAuthorizedKeys = %q, want %q", got, want)
	}
}

func TestConfigExportError(t *testing.T) {
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		return dockertest.ExecResponse{ExitCode: 1}
	})

	b := New(d.Client)
	if _, err := ConfigExport(b); err == nil {
		t.Fatalf("ConfigExport: got nil error, want error on non-zero exit")
	}
}

func TestConfigGet(t *testing.T) {
	var gotCmd []string
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		gotCmd = cmd
		// The trailing newline must be trimmed by ConfigGet.
		return dockertest.ExecResponse{Stdout: "2222\n"}
	})

	b := New(d.Client)
	got, err := ConfigGet(b, "rancher.ssh.port")
	if err != nil {
		t.Fatalf("ConfigGet: %v", err)
	}

	if want := []string{
		"ros", "config", "get", "rancher.ssh.port",
	}; !reflect.DeepEqual(gotCmd, want) {
		t.Errorf("exec cmd = %q, want %q", gotCmd, want)
	}
	if got != "2222" {
		t.Errorf("ConfigGet = %q, want %q", got, "2222")
	}
}

func TestConfigGetError(t *testing.T) {
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		return dockertest.ExecResponse{ExitCode: 1}
	})

	b := New(d.Client)
	if _, err := ConfigGet(b, "rancher.ssh.port"); err == nil {
		t.Fatalf("ConfigGet: got nil error, want error on non-zero exit")
	}
}

func TestConfigSet(t *testing.T) {
	var gotCmd []string
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		gotCmd = cmd
		return dockertest.ExecResponse{}
	})

	b := New(d.Client)
	if err := ConfigSet(b, "rancher.ssh.port", "2222"); err != nil {
		t.Fatalf("ConfigSet: %v", err)
	}

	want := []string{"ros", "config", "set", "rancher.ssh.port", "2222"}
	if !reflect.DeepEqual(gotCmd, want) {
		t.Errorf("exec cmd = %q, want %q", gotCmd, want)
	}
}

func TestConfigSetError(t *testing.T) {
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		return dockertest.ExecResponse{ExitCode: 1}
	})

	b := New(d.Client)
	if err := ConfigSet(b, "k", "v"); err == nil {
		t.Fatalf("ConfigSet: got nil error, want error on non-zero exit")
	}
}

func TestConfigMerge(t *testing.T) {
	var script string
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		if len(cmd) == 3 && cmd[0] == "/bin/bash" && cmd[1] == "-c" {
			script = cmd[2]
		}
		return dockertest.ExecResponse{}
	})

	b := New(d.Client)
	config := &bosinit.Config{SSHAuthorizedKeys: []string{"ssh-rsa key1"}}
	if err := ConfigMerge(b, config); err != nil {
		t.Fatalf("ConfigMerge: %v", err)
	}

	for _, want := range []string{
		"sudo ros config merge", // pipes the cloud config into ros
		"#cloud-config",         // the encoded config keeps its shebang
		"ssh-rsa key1",          // the key we asked to merge
	} {
		if !strings.Contains(script, want) {
			t.Errorf("merge script %q missing %q", script, want)
		}
	}
}

func TestConfigMergeError(t *testing.T) {
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		return dockertest.ExecResponse{ExitCode: 1}
	})

	b := New(d.Client)
	config := &bosinit.Config{SSHAuthorizedKeys: []string{"ssh-rsa key1"}}
	if err := ConfigMerge(b, config); err == nil {
		t.Fatalf("ConfigMerge: got nil error, want error on non-zero exit")
	}
}

func TestAddSSHKeys(t *testing.T) {
	// The console already has one key configured.
	const exported = "ssh_authorized_keys:\n- ssh-rsa existing\n"

	var script string
	d := newExecDaemon(t, func(cmd []string) dockertest.ExecResponse {
		if len(cmd) == 3 && cmd[0] == "ros" && cmd[1] == "config" &&
			cmd[2] == "export" {
			return dockertest.ExecResponse{Stdout: exported}
		}
		if len(cmd) == 3 && cmd[0] == "/bin/bash" && cmd[1] == "-c" {
			script = cmd[2]
		}
		return dockertest.ExecResponse{}
	})

	b := New(d.Client)
	// "ssh-rsa existing" is already present and must be deduplicated;
	// "ssh-rsa new" must be added.
	if err := AddSSHKeys(b, []string{"ssh-rsa existing", "ssh-rsa new"}); err != nil {
		t.Fatalf("AddSSHKeys: %v", err)
	}

	if !strings.Contains(script, "ssh-rsa new") {
		t.Errorf("merge script %q missing newly added key", script)
	}
	if n := strings.Count(script, "ssh-rsa existing"); n != 1 {
		t.Errorf(
			"existing key appears %d times in merge script, want 1: %q",
			n, script,
		)
	}
}
