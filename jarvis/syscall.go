//go:build !linux

package jarvis

import (
	"errors"
)

var errNotLinux = errors.New("not linux; not implemented")

func mountBootPart(dev, path string) error {
	return errNotLinux
}

func unmountBootPart(path string) error {
	return errNotLinux
}
