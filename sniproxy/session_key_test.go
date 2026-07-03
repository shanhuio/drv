package sniproxy

import "testing"

func TestSessionKeyRoundTrip(t *testing.T) {
	for _, k := range []sessionKey{
		{},
		{ID: 1, Key: 2},
		{ID: 0xffffffffffffffff, Key: 0xdeadbeefcafef00d},
	} {
		s, err := k.encode()
		if err != nil {
			t.Errorf("encode %+v: %s", k, err)
			continue
		}
		got, err := decodeSessionKey(s)
		if err != nil {
			t.Errorf("decode %q: %s", s, err)
			continue
		}
		if *got != k {
			t.Errorf("round-trip %+v: got %+v", k, *got)
		}
	}
}

func TestDecodeSessionKeyError(t *testing.T) {
	for _, s := range []string{
		"",
		"not json",
		"{",
		`{"ID":"nope"}`, // wrong type for uint64 field
	} {
		if got, err := decodeSessionKey(s); err == nil {
			t.Errorf("decodeSessionKey(%q): got %+v, want error", s, got)
		}
	}
}
