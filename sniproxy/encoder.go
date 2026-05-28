package sniproxy

import (
	"io"
)

type encoder struct {
	w   io.Writer
	n   int64
	err error
}

func newEncoder(w io.Writer) *encoder {
	return &encoder{w: w}
}

func (e *encoder) write(bs []byte) {
	if e.hasErr() {
		return
	}
	n, err := e.w.Write(bs)
	if err != nil {
		e.err = err
		return
	}
	e.n += int64(n)
}

func (e *encoder) Err() error { return e.err }

func (e *encoder) hasErr() bool { return e.err != nil }

func (e *encoder) u64(v uint64) {
	if e.hasErr() {
		return // To shortcut the encoding part.
	}
	var bs [8]byte
	endian.PutUint64(bs[:], v)
	e.write(bs[:])
}

func (e *encoder) u8(v uint8) {
	if e.hasErr() {
		return
	}
	e.write([]byte{v})
}

func (e *encoder) bytes(bs []byte) {
	e.u64(uint64(len(bs)))
	e.write(bs)
}

func (e *encoder) str(s string) {
	bs := []byte(s)
	e.bytes(bs)
}
