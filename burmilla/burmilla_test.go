package burmilla

import (
	"testing"

	"shanhu.io/std/docker"
	"shanhu.io/std/docker/dockertest"
	"shanhu.io/std/tarutil"
)

// newConsoleDaemon starts a fake docker daemon that already has the
// "console" container the Burmilla stub talks to. It registers cleanup that
// closes the daemon and surfaces any internal errors it recorded.
func newConsoleDaemon(t *testing.T) *dockertest.FakeDaemon {
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
	d.AddContainer(&dockertest.Container{ID: "console"})
	return d
}

func TestConsoleName(t *testing.T) {
	d := newConsoleDaemon(t)
	b := New(d.Client)
	if got := b.Console().ID(); got != "console" {
		t.Errorf("Console().ID() = %q, want %q", got, "console")
	}
}

func TestExecOutput(t *testing.T) {
	d := newConsoleDaemon(t)
	d.SetExecResponse(dockertest.ExecResponse{Stdout: "hello\n", ExitCode: 0})

	b := New(d.Client)
	out, err := b.ExecOutput([]string{"echo", "hello"})
	if err != nil {
		t.Fatalf("ExecOutput: %v", err)
	}
	if string(out) != "hello\n" {
		t.Errorf("ExecOutput = %q, want %q", out, "hello\n")
	}
}

func TestExecOutputError(t *testing.T) {
	d := newConsoleDaemon(t)
	d.SetExecResponse(dockertest.ExecResponse{Stdout: "partial", ExitCode: 1})

	b := New(d.Client)
	out, err := b.ExecOutput([]string{"false"})
	if err == nil {
		t.Fatalf("ExecOutput: got nil error, want non-zero exit error")
	}
	if out != nil {
		t.Errorf("ExecOutput output = %q, want nil on error", out)
	}
}

func TestExecRet(t *testing.T) {
	for _, want := range []int{0, 1, 42} {
		d := newConsoleDaemon(t)
		d.SetExecResponse(dockertest.ExecResponse{ExitCode: want})

		b := New(d.Client)
		got, err := b.ExecRet([]string{"some", "command"})
		if err != nil {
			t.Fatalf("ExecRet: %v", err)
		}
		if got != want {
			t.Errorf("ExecRet = %d, want %d", got, want)
		}
	}
}

func TestCopyInTarStream(t *testing.T) {
	d := newConsoleDaemon(t)
	b := New(d.Client)

	const payload = "the file content"
	s := tarutil.NewStream()
	s.AddString("data.txt", tarutil.ModeMeta(0644), payload)

	if err := b.CopyInTarStream(s, "/opt"); err != nil {
		t.Fatalf("CopyInTarStream: %v", err)
	}

	bs, err := docker.ReadContFile(docker.NewCont(d.Client, "console"), "/opt/data.txt")
	if err != nil {
		t.Fatalf("read back copied file: %v", err)
	}
	if string(bs) != payload {
		t.Errorf("copied content = %q, want %q", bs, payload)
	}
}

func TestListOS(t *testing.T) {
	d := newConsoleDaemon(t)
	// Output mixes blank lines and surrounding whitespace, which ListOS
	// must trim and drop.
	d.SetExecResponse(dockertest.ExecResponse{
		Stdout: "\n  v1.0.0 \nv1.1.0\n\n\tv1.2.0\t\n\n",
	})

	b := New(d.Client)
	got, err := ListOS(b)
	if err != nil {
		t.Fatalf("ListOS: %v", err)
	}
	want := []string{"v1.0.0", "v1.1.0", "v1.2.0"}
	if len(got) != len(want) {
		t.Fatalf("ListOS = %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("ListOS[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestListOSError(t *testing.T) {
	d := newConsoleDaemon(t)
	d.SetExecResponse(dockertest.ExecResponse{ExitCode: 1})

	b := New(d.Client)
	if _, err := ListOS(b); err == nil {
		t.Fatalf("ListOS: got nil error, want error from non-zero exit")
	}
}
