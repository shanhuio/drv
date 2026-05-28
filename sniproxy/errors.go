package sniproxy

import (
	"errors"
)

var (
	errAlreadyShutdown = errors.New("already shutdown")
	errAlreadyClosed   = errors.New("already closed")
)
