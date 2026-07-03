package sniproxy

import "testing"

func TestDialRequestRoundTrip(t *testing.T) {
	roundTrip(t, &dialRequest{}, new(dialRequest))
}

func TestDialResponseRoundTrip(t *testing.T) {
	for _, m := range []*dialResponse{
		{},
		{session: 7},
		{session: 7, err: newRemoteErrString(errAccept, "not accepting")},
	} {
		roundTrip(t, m, new(dialResponse))
	}
}

func TestDialSideRequestRoundTrip(t *testing.T) {
	for _, m := range []*dialSideRequest{
		{},
		{session: 1, key: 2, token: "tok"},
	} {
		roundTrip(t, m, new(dialSideRequest))
	}
}

func TestDialSide2RequestRoundTrip(t *testing.T) {
	for _, m := range []*dialSide2Request{
		{},
		{session: 1, key: 2, token: "tok", tcpAddr: "10.0.0.1:443"},
	} {
		roundTrip(t, m, new(dialSide2Request))
	}
}
