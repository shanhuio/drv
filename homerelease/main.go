package homerelease

import (
	"shanhu.io/g/subcmd"
)

func cmd() *subcmd.List {
	c := subcmd.New()
	c.Add("build", "build a release", cmdBuild)
	c.AddHost("push", "pushes a release", cmdPush)
	return c
}

// Main is the main entrance function.
func Main() { cmd().Main() }
