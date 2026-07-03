package sniproxy

import "testing"

func TestCloseRequestRoundTrip(t *testing.T) {
	for _, m := range []*closeRequest{
		{},
		{session: 42},
	} {
		roundTrip(t, m, new(closeRequest))
	}
}

func TestCloseResponseRoundTrip(t *testing.T) {
	for _, m := range []*closeResponse{
		{},
		{err: newRemoteErrString(errAccept, "not accepting")},
	} {
		roundTrip(t, m, new(closeResponse))
	}
}
