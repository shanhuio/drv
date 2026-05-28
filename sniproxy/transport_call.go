package sniproxy

import (
	"context"
)

type transportCall struct {
	context context.Context
	typ     byte
	req     encoderTo
	resp    decoderFrom
	done    func(err error)
}

func newTransportCall(
	ctx context.Context, t byte, req encoderTo, resp decoderFrom,
) *transportCall {
	return &transportCall{
		context: ctx,
		typ:     t,
		req:     req,
		resp:    resp,
	}
}
