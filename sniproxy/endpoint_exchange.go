package sniproxy

import (
	"github.com/gorilla/websocket"
)

type endpointExchange struct {
	id      uint64
	t       uint8
	errcode uint8
	req     decoderFrom
	resp    encoderTo
}

func (x *endpointExchange) encodeTo(enc *encoder) {
	enc.u64(x.id)
	enc.u8(x.t)
	enc.u8(x.errcode)

	if x.resp != nil {
		x.resp.encodeTo(enc)
	}
}

func writeExchangeResp(conn *websocket.Conn, x *endpointExchange) error {
	w, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}

	enc := newEncoder(w)
	x.encodeTo(enc)
	if enc.hasErr() {
		return enc.Err()
	}
	return w.Close()
}
