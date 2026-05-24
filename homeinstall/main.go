package homeinstall

import (
	"shanhu.io/g/subcmd"
)

// Main is the main entrance for the command line.
func Main() {
	c := subcmd.New()
	c.Add(
		"config", "prompts for config and generates install script",
		cmdConfig,
	)
	c.Main()
}
