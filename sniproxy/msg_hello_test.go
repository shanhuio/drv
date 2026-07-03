package sniproxy

import "testing"

func TestHelloRequestRoundTrip(t *testing.T) {
	for _, m := range []*helloRequest{
		{},
		{msg: "hello"},
	} {
		roundTrip(t, m, new(helloRequest))
	}
}

func TestHelloResponseRoundTrip(t *testing.T) {
	for _, m := range []*helloResponse{
		{},
		{msg: "world"},
	} {
		roundTrip(t, m, new(helloResponse))
	}
}
