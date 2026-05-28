package sniproxy

var (
	_ = []message{
		new(dialRequest),
		new(dialResponse),
		new(closeRequest),
		new(closeResponse),
		new(readRequest),
		new(readResponse),
		new(writeRequest),
		new(writeResponse),
		new(statusRequest),
		new(statusResponse),
	}
)
