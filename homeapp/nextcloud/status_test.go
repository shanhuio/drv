package nextcloud

import (
	"testing"
)

func TestParseNextcloudStatus(t *testing.T) {
	const s = "Nextcloud is weird...\n" +
		`{"installed": false, "version": "0.1"}`

	status, err := parseNextcloudStatus(s)
	if err != nil {
		t.Errorf("parse %q: %s", s, err)
	}
	if status.Installed != false {
		t.Error("want not installed")
	}
	if want := "0.1"; status.Version != want {
		t.Errorf("wrong version: got %q, want %q", status.Version, want)
	}
}
