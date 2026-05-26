package apputil

import (
	"errors"
	"log"

	"shanhu.io/g/dock"
	"shanhu.io/std/errcode"
)

// ErrSameImage is returned when there is no image change.
var ErrSameImage = errors.New("same image")

// DropIfDifferent drops the container with name if the image is
// different. It returns ErrSameImage if the image is the same.
func DropIfDifferent(d *dock.Client, name, img string) error {
	c := dock.NewCont(d, name)
	info, err := c.Inspect()
	if err != nil {
		if errcode.IsNotFound(err) {
			log.Printf("container %q not found", name)
			return nil
		}
		return errcode.Annotatef(err, "inspect %s", name)
	}
	if info.Image == img {
		return ErrSameImage // nothing to update
	}
	if err := c.Drop(); err != nil {
		return errcode.Annotatef(err, "drop %s", name)
	}
	return nil
}

// DropIfExists drops the container if the container exists.  Otherwise, it
// prints a log line and do nothing.
func DropIfExists(cont *dock.Cont) error {
	exists, err := cont.Exists()
	if err != nil {
		return errcode.Annotatef(err, "check container exists")
	}
	if !exists {
		log.Printf("container %q does not exist; skip dropping", cont.ID())
		return nil
	}
	return cont.Drop()
}
