package sniproxy

import "testing"

func TestStatusRequestRoundTrip(t *testing.T) {
	for _, m := range []*statusRequest{
		{},
		{session: 9},
	} {
		roundTrip(t, m, new(statusRequest))
	}
}

func TestStatusResponseRoundTrip(t *testing.T) {
	for _, m := range []*statusResponse{
		{},
		{uptime: 100, totalRead: 200, totalWritten: 300},
	} {
		roundTrip(t, m, new(statusResponse))
	}
}
