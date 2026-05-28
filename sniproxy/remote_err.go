package sniproxy

import (
	"io"
)

type remoteErr struct {
	code    int
	message string
}

func newRemoteErr(code int, err error) *remoteErr {
	return newRemoteErrString(code, err.Error())
}

func newRemoteErrString(code int, s string) *remoteErr {
	return &remoteErr{
		code:    code,
		message: s,
	}
}

func (e *remoteErr) Error() string { return e.message }

func (e *remoteErr) toError() error {
	if e == nil {
		return nil
	}
	if e.code == errEOF {
		return io.EOF
	}
	return e
}

func encodeRemoteErr(enc *encoder, err *remoteErr) {
	if err == nil {
		var empty remoteErr
		empty.encodeTo(enc)
		return
	}
	err.encodeTo(enc)
}

func (e *remoteErr) encodeTo(enc *encoder) {
	if e.code == 0 {
		enc.u64(0)
		return
	}
	enc.u64(uint64(e.code))
	enc.str(e.message)
}

func (e *remoteErr) decodeFrom(dec *decoder) {
	e.code = int(dec.u64())
	if e.code != 0 {
		e.message = dec.str()
	}
}

func decodeRemoteErr(dec *decoder) *remoteErr {
	err := new(remoteErr)
	err.decodeFrom(dec)
	if err.code == 0 {
		return nil
	}
	return err
}

var (
	remoteErrNotAccepting = newRemoteErrString(errAccept, "not accepting")
	remoteErrSiding       = newRemoteErrString(errSiding, "tunnel is siding")
)
