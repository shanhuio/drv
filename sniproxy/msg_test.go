package sniproxy

import (
	"bytes"
	"reflect"
	"testing"
)

// roundTrip encodes in, decodes the bytes back into out, and checks that out
// equals in. out must be a fresh, zero-valued pointer of the same concrete
// type as in.
func roundTrip(t *testing.T, in, out message) {
	t.Helper()

	buf := new(bytes.Buffer)
	enc := newEncoder(buf)
	in.encodeTo(enc)
	if err := enc.Err(); err != nil {
		t.Fatalf("encode %T: %s", in, err)
	}

	dec := newDecoder(buf)
	out.decodeFrom(dec)
	dec.end()
	if err := dec.Err(); err != nil {
		t.Fatalf("decode %T: %s", in, err)
	}

	if !reflect.DeepEqual(in, out) {
		t.Errorf("round-trip %T: got %+v, want %+v", in, out, in)
	}
}
