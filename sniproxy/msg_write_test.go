package sniproxy

import "testing"

func TestWriteRequestRoundTrip(t *testing.T) {
	for _, m := range []*writeRequest{
		{},
		{session: 5, bytes: []byte{0x01, 0x02, 0x03}},
	} {
		roundTrip(t, m, new(writeRequest))
	}
}

func TestWriteResponseRoundTrip(t *testing.T) {
	for _, m := range []*writeResponse{
		{},
		{written: 128},
		{written: 0, err: newRemoteErrString(errWrite, "write failed")},
	} {
		roundTrip(t, m, new(writeResponse))
	}
}
