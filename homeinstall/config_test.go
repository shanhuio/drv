package homeinstall

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCheckPassCode(t *testing.T) {
	for _, test := range []struct {
		code    string
		wantErr bool
	}{
		{code: "0", wantErr: false},
		{code: "123456", wantErr: false},
		{code: "0000000000", wantErr: false},
		{code: "", wantErr: true},       // empty is rejected
		{code: "12345a", wantErr: true}, // letters are rejected
		{code: "12 34", wantErr: true},  // spaces are rejected
		{code: "-123", wantErr: true},   // punctuation is rejected
		{code: "１２３", wantErr: true},    // non-ASCII digits are rejected
	} {
		err := checkPassCode(test.code)
		if test.wantErr && err == nil {
			t.Errorf("checkPassCode(%q): got nil error, want error", test.code)
		}
		if !test.wantErr && err != nil {
			t.Errorf("checkPassCode(%q): got error %v, want nil", test.code, err)
		}
	}
}

func TestCheckEndpointName(t *testing.T) {
	for _, test := range []struct {
		name    string
		wantErr bool
	}{
		{name: "a", wantErr: false},
		{name: "my-endpoint-1", wantErr: false},
		{name: "0123456789", wantErr: false},
		{name: "abcdefghijklmnopqrstuvwxyz", wantErr: false},
		{name: "", wantErr: true},           // empty is rejected
		{name: "MyEndpoint", wantErr: true}, // upper case is rejected
		{name: "end point", wantErr: true},  // spaces are rejected
		{name: "end_point", wantErr: true},  // underscores are rejected
		{name: "end.point", wantErr: true},  // dots are rejected
		{name: "café", wantErr: true},       // non-ASCII letters are rejected
	} {
		err := checkEndpointName(test.name)
		if test.wantErr && err == nil {
			t.Errorf(
				"checkEndpointName(%q): got nil error, want error", test.name,
			)
		}
		if !test.wantErr && err != nil {
			t.Errorf(
				"checkEndpointName(%q): got error %v, want nil", test.name, err,
			)
		}
	}
}

func TestInstallScriptArgs(t *testing.T) {
	s := &installScript{
		endpoint: "my-endpoint",
		passCode: "123456",
		bin:      "/opt/homedrv/install",
	}
	got := s.args()
	want := []string{
		"sudo", "/opt/homedrv/install",
		"--name", "my-endpoint",
		"--code", "123456",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("args() = %q, want %q", got, want)
	}
}

func TestInstallScriptBashScript(t *testing.T) {
	s := &installScript{
		endpoint: "my-endpoint",
		passCode: "123456",
		bin:      "/opt/homedrv/install",
	}
	got := string(s.bashScript())
	want := "#/bin/bash\n" +
		"sudo /opt/homedrv/install --name my-endpoint --code 123456\n"
	if got != want {
		t.Errorf("bashScript() =\n%q\nwant\n%q", got, want)
	}
}

func TestInstallScriptWriteOut(t *testing.T) {
	s := &installScript{
		endpoint: "my-endpoint",
		passCode: "123456",
		bin:      "/opt/homedrv/install",
	}

	f := filepath.Join(t.TempDir(), "install.sh")
	if err := s.writeOut(f); err != nil {
		t.Fatalf("writeOut(%q): %v", f, err)
	}

	bs, err := os.ReadFile(f)
	if err != nil {
		t.Fatalf("read back %q: %v", f, err)
	}
	if got, want := string(bs), string(s.bashScript()); got != want {
		t.Errorf("written file =\n%q\nwant\n%q", got, want)
	}

	info, err := os.Stat(f)
	if err != nil {
		t.Fatalf("stat %q: %v", f, err)
	}
	if perm := info.Mode().Perm(); perm != 0755 {
		t.Errorf("written file perm = %o, want %o", perm, 0755)
	}
}
