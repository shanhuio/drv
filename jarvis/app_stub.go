package jarvis

import (
	"shanhu.io/homedrv/drv/homeapp"
)

type appMaker interface {
	makeStub(name string) (*appStub, error)
}

type appStub struct {
	homeapp.App
}
