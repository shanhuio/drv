package jarvis

import (
	"shanhu.io/std/errcode"
)

func cmdUpdateGrubConfig(args []string) error {
	flags := cmdFlags.New()
	dev := flags.String("dev", "/dev/sda1", "boot partition device")
	osVersion := flags.String("os", "burmilla/os:v1.9.1", "os version")
	args = flags.ParseArgs(args)
	if len(args) != 0 {
		return errcode.InvalidArgf("expects no args")
	}
	return updateBootPart(*dev, *osVersion)
}
