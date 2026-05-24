// Package homeboot provides a command line tool that creates a private key
// identity and enrolls it.
package homeboot

import (
	"shanhu.io/g/subcmd"
)

// Main is the main entrance of the command line.
func Main() {
	c := subcmd.New()
	c.Add("install", "installs a new homedrive", cmdInstall)
	c.Add("uninstall", "uninstalls homedrive installation", cmdUninstall)
	c.Add(
		"cloud-config", "prints cloud-config for a new homedrive",
		cmdCloudConfig,
	)
	c.Add("enroll", "manually enroll an endpoint using passcode", cmdEnroll)

	c.Main()
}
