package jarvis

import (
	"shanhu.io/g/errcode"
	"shanhu.io/g/flagutil"
	"shanhu.io/g/osutil"
)

var cmdFlags = flagutil.NewFactory("jarvis")

type clientFlags struct {
	home string
}

func newClientFlags(flags *flagutil.FlagSet) *clientFlags {
	c := new(clientFlags)
	flags.StringVar(&c.home, "home", ".", "home directory")
	return c
}

func newClientDrive(flags *clientFlags) (*drive, error) {
	h, err := osutil.NewHome(flags.home)
	if err != nil {
		return nil, errcode.Annotate(err, "new home")
	}
	c, err := readConfig(h)
	if err != nil {
		return nil, errcode.Annotate(err, "read config")
	}
	b, err := newBackend(h)
	if err != nil {
		return nil, err
	}
	d, err := newDrive(c, b.kernel())
	if err != nil {
		return nil, err
	}
	return d, nil
}
