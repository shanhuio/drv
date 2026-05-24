package drvapi

import (
	"testing"
)

func TestChannelName(t *testing.T) {
	for _, test := range []struct {
		name string
		base string
		arch string
	}{
		{"stable", "stable", "amd64"},
		{"alpha", "alpha", "amd64"},
		{"stable-amd64", "stable", "amd64"},
		{"stable-arm64", "stable", "arm64"},
		{"alpha-arm64", "alpha", "arm64"},
	} {
		parsed := ParseChannelName(test.name)
		if parsed.Base != test.base {
			t.Errorf(
				"parse channel name: got base %q, want %q",
				parsed.Base, test.base,
			)
		}
		if got := parsed.Architecture(); got != test.arch {
			t.Errorf(
				"parse channel name: got arch %q, want %q",
				got, test.arch,
			)
		}
		if got := parsed.String(); got != test.name {
			t.Errorf(
				"parsed channel name: got %q, want %q",
				got, test.name,
			)
		}
	}
}
