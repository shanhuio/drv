package sniproxy

type readRequest struct {
	session uint64
	maxRead int
}

func (m *readRequest) encodeTo(enc *encoder) {
	enc.u64(m.session)
	enc.u64(uint64(m.maxRead))
}

func (m *readRequest) decodeFrom(dec *decoder) {
	m.session = dec.u64()
	m.maxRead = int(dec.u64())
}

type readResponse struct {
	bytes []byte
	err   *remoteErr
}

func (m *readResponse) encodeTo(enc *encoder) {
	enc.bytes(m.bytes)
	encodeRemoteErr(enc, m.err)
}

func (m *readResponse) decodeFrom(dec *decoder) {
	m.bytes = dec.bytes(m.bytes)
	m.err = decodeRemoteErr(dec)
}
