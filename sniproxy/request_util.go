package sniproxy

func newRequestMessage(t uint8) (decoderFrom, bool) {
	switch t {
	case msgShutdown:
		return nil, true
	case msgHello:
		return new(helloRequest), true
	case msgDial:
		return new(dialRequest), true
	case msgDialSide:
		return new(dialSideRequest), true
	case msgDialSide2:
		return new(dialSide2Request), true
	case msgRead:
		return new(readRequest), true
	case msgWrite:
		return new(writeRequest), true
	case msgClose:
		return new(closeRequest), true
	default:
		return nil, false
	}
}
