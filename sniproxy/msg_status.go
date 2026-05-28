package sniproxy

type statusRequest struct {
	session uint64
}

func (m *statusRequest) encodeTo(enc *encoder) {
	enc.u64(m.session)
}

func (m *statusRequest) decodeFrom(dec *decoder) {
	m.session = dec.u64()
}

type statusResponse struct {
	uptime       uint64
	totalRead    uint64
	totalWritten uint64
}

func (m *statusResponse) encodeTo(enc *encoder) {
	enc.u64(m.uptime)
	enc.u64(m.totalRead)
	enc.u64(m.totalWritten)
}

func (m *statusResponse) decodeFrom(dec *decoder) {
	m.uptime = dec.u64()
	m.totalRead = dec.u64()
	m.totalWritten = dec.u64()
}
