package bosinit

import (
	"testing"

	"reflect"
	"strings"
)

func TestParseConfig(t *testing.T) {
	content := strings.Join([]string{
		"ssh_authorized_keys:",
		"- ssh-rsa key1",
		"- ssh-rsa key2",
	}, "\n")

	config, err := ParseConfig([]byte(content))
	if err != nil {
		t.Fatal("parse config: ", err)
	}

	wantKeys := []string{
		"ssh-rsa key1",
		"ssh-rsa key2",
	}
	if got := config.SSHAuthorizedKeys; !reflect.DeepEqual(got, wantKeys) {
		t.Errorf("parse config, got keys %q, want %q", got, wantKeys)
	}
}

func TestParseConfigRancher(t *testing.T) {
	content := strings.Join([]string{
		"rancher:",
		"  ssh:",
		"    port: 2222",
		"  upgrade:",
		"    url: https://example.com/upgrade",
		"  cloud_init:",
		"    datasources:",
		"    - configdrive",
		"  resize_device: /dev/sda",
		"write_files:",
		"- path: /etc/rc.local",
		"  permissions: \"0744\"",
		"  owner: root",
		"  content: echo hi",
	}, "\n")

	config, err := ParseConfig([]byte(content))
	if err != nil {
		t.Fatal("parse config: ", err)
	}

	want := &Config{
		Rancher: &Rancher{
			SSH:          &SSH{Port: 2222},
			Upgrade:      &Upgrade{URL: "https://example.com/upgrade"},
			CloudInit:    &CloudInit{DataSources: []string{"configdrive"}},
			ResizeDevice: "/dev/sda",
		},
		WriteFiles: []*WriteFile{{
			Path:        "/etc/rc.local",
			Permissions: "0744",
			Owner:       "root",
			Content:     "echo hi",
		}},
	}
	if !reflect.DeepEqual(config, want) {
		t.Errorf("parse config:\n got %+v\nwant %+v", config, want)
	}
}

func TestParseConfigInvalid(t *testing.T) {
	// A scalar where a mapping is expected is not a valid Config.
	if _, err := ParseConfig([]byte("just a string")); err == nil {
		t.Errorf("ParseConfig: got nil error, want error on invalid YAML")
	}
}

func TestCloudConfigShebang(t *testing.T) {
	c := &Config{SSHAuthorizedKeys: []string{"ssh-rsa key1"}}
	bs, err := c.CloudConfig()
	if err != nil {
		t.Fatal("cloud config: ", err)
	}

	const shebang = "#cloud-config\n"
	if !strings.HasPrefix(string(bs), shebang) {
		t.Errorf("CloudConfig missing shebang, got:\n%s", bs)
	}

	// The remainder after the shebang must parse back into the same config.
	rest := strings.TrimPrefix(string(bs), shebang)
	got, err := ParseConfig([]byte(rest))
	if err != nil {
		t.Fatal("parse cloud config body: ", err)
	}
	if !reflect.DeepEqual(got, c) {
		t.Errorf("round trip:\n got %+v\nwant %+v", got, c)
	}
}

func TestEncodeNoShebang(t *testing.T) {
	c := &Config{SSHAuthorizedKeys: []string{"ssh-rsa key1"}}
	bs, err := c.Encode()
	if err != nil {
		t.Fatal("encode: ", err)
	}
	if strings.HasPrefix(string(bs), "#cloud-config") {
		t.Errorf("Encode must not include shebang, got:\n%s", bs)
	}
}

func TestEncodeRoundTrip(t *testing.T) {
	want := &Config{
		Rancher: &Rancher{
			SSH:          &SSH{Port: 22},
			Upgrade:      &Upgrade{URL: "https://example.com/u"},
			CloudInit:    &CloudInit{DataSources: []string{"ec2", "configdrive"}},
			ResizeDevice: "/dev/vda",
		},
		WriteFiles: []*WriteFile{
			RCLocal("echo hello"),
			{Path: "/tmp/x", Permissions: FilePerm(0600), Owner: "rancher", Content: "x"},
		},
		SSHAuthorizedKeys: []string{"ssh-rsa key1", "ssh-rsa key2"},
	}

	bs, err := want.Encode()
	if err != nil {
		t.Fatal("encode: ", err)
	}
	got, err := ParseConfig(bs)
	if err != nil {
		t.Fatal("parse: ", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("round trip:\n got %+v\nwant %+v", got, want)
	}
}

func TestEncodeOmitEmpty(t *testing.T) {
	bs, err := (&Config{}).Encode()
	if err != nil {
		t.Fatal("encode: ", err)
	}
	for _, key := range []string{"rancher", "write_files", "ssh_authorized_keys"} {
		if strings.Contains(string(bs), key) {
			t.Errorf("empty config encoded %q, want it omitted; got:\n%s", key, bs)
		}
	}
}

func TestEncodeWriteFilesKey(t *testing.T) {
	// The YAML key must be "write_files", not the default "writefiles".
	c := &Config{WriteFiles: []*WriteFile{RCLocal("x")}}
	bs, err := c.Encode()
	if err != nil {
		t.Fatal("encode: ", err)
	}
	if !strings.Contains(string(bs), "write_files:") {
		t.Errorf("encoded config missing write_files key, got:\n%s", bs)
	}
}
