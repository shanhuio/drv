package sniproxy

type writeRequest struct {
	session uint64
	bytes   []byte
}

func (m *writeRequest) encodeTo(enc *encoder) {
	enc.u64(m.session)
	enc.bytes(m.bytes)
}

func (m *writeRequest) decodeFrom(dec *decoder) {
	m.session = dec.u64()
	m.bytes = dec.bytes(m.bytes)
}

type writeResponse struct {
	written int
	err     *remoteErr
}

func (m *writeResponse) encodeTo(enc *encoder) {
	enc.u64(uint64(m.written))
	encodeRemoteErr(enc, m.err)
}

func (m *writeResponse) decodeFrom(dec *decoder) {
	m.written = int(dec.u64())
	m.err = decodeRemoteErr(dec)
}
