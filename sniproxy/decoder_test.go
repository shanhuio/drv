package sniproxy

import (
	"testing"

	"bytes"
	"math/rand"
	"reflect"
)

func TestDecoder(t *testing.T) {
	bigTrunk := make([]byte, 1024*32)
	if _, err := rand.Read(bigTrunk); err != nil {
		t.Fatal("generate big trunk data", err)
	}

	type testCase struct {
		en   func(enc *encoder)
		de   func(dec *decoder) any
		want any
	}

	for _, c := range []*testCase{{
		en:   func(enc *encoder) { enc.u64(32) },
		de:   func(dec *decoder) any { return dec.u64() },
		want: uint64(32),
	}, {
		en:   func(enc *encoder) { enc.str("hello") },
		de:   func(dec *decoder) any { return dec.str() },
		want: "hello",
	}, {
		en:   func(enc *encoder) { enc.u8(3) },
		de:   func(dec *decoder) any { return dec.u8() },
		want: uint8(3),
	}, {
		en:   func(enc *encoder) { enc.bytes([]byte{0xfa, 0xce}) },
		de:   func(dec *decoder) any { return dec.bytes(nil) },
		want: []byte{0xfa, 0xce},
	}, {
		en:   func(enc *encoder) { enc.bytes(bigTrunk) },
		de:   func(dec *decoder) any { return dec.bytes(nil) },
		want: bigTrunk,
	}} {
		buf := new(bytes.Buffer)
		enc := newEncoder(buf)
		c.en(enc)
		if err := enc.Err(); err != nil {
			t.Errorf("encode got error: %s", err)
		}

		dec := newDecoder(buf)
		got := c.de(dec)
		dec.end()

		if err := dec.Err(); err != nil {
			t.Errorf("decode got error: %s", err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("want %v, got %v", c.want, got)
		}
	}

}
