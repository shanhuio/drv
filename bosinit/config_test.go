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
