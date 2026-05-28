package sniproxy

type closeRequest struct {
	session uint64
}

func (m *closeRequest) encodeTo(enc *encoder) {
	enc.u64(m.session)
}

func (m *closeRequest) decodeFrom(dec *decoder) {
	m.session = dec.u64()
}

type closeResponse struct {
	err *remoteErr
}

func (m *closeResponse) encodeTo(enc *encoder) {
	encodeRemoteErr(enc, m.err)
}

func (m *closeResponse) decodeFrom(dec *decoder) {
	m.err = decodeRemoteErr(dec)
}
