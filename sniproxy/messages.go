package sniproxy

// Message types. For backwards compatibility, new message types must be added
// in the back.
const (
	msgShutdown = iota
	msgHello
	msgDial
	msgWrite
	msgRead
	msgStatus
	msgClose
	msgShutdownHint

	msgDialSide
	msgDialSide2
)

type encoderTo interface {
	encodeTo(enc *encoder)
}

type decoderFrom interface {
	decodeFrom(dec *decoder)
}

type message interface {
	encoderTo
	decoderFrom
}

const (
	errUnknown = iota + 1
	errUnknownType
	errBug
	errAccept
	errSessionNotFound
	errRead
	errWrite
	errClose
	errInternal
	errEOF
	errSiding
)
