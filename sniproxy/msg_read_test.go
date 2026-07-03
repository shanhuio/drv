package sniproxy

import "testing"

func TestReadRequestRoundTrip(t *testing.T) {
	for _, m := range []*readRequest{
		{},
		{session: 3, maxRead: 4096},
	} {
		roundTrip(t, m, new(readRequest))
	}
}

func TestReadResponseRoundTrip(t *testing.T) {
	for _, m := range []*readResponse{
		{},
		{bytes: []byte{0xde, 0xad, 0xbe, 0xef}},
		{err: newRemoteErrString(errRead, "read failed")},
		{bytes: []byte("payload"), err: newRemoteErrString(errEOF, "eof")},
	} {
		roundTrip(t, m, new(readResponse))
	}
}
