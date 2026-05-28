package sniproxy

type helloRequest struct {
	msg string
}

func (m *helloRequest) encodeTo(enc *encoder) {
	enc.str(m.msg)
}

func (m *helloRequest) decodeFrom(dec *decoder) {
	m.msg = dec.str()
}

type helloResponse struct {
	msg string
}

func (m *helloResponse) encodeTo(enc *encoder) {
	enc.str(m.msg)
}

func (m *helloResponse) decodeFrom(dec *decoder) {
	m.msg = dec.str()
}
