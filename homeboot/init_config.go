package homeboot

import (
	"shanhu.io/g/flagutil"
)

// InitConfig is the configuration to initialize a homedrive.
type InitConfig struct {
	HomeBoot string

	Boot *BootConfig

	GitHubKeys string
	UserKeys   string
}

// NewInitConfig creates a new init config for creating a new
// HomeDrive instance.
func NewInitConfig() *InitConfig {
	return &InitConfig{Boot: newBootConfig()}
}

// DeclareFlags declares command line flags.
func (c *InitConfig) DeclareFlags(flags *flagutil.FlagSet) {
	c.Boot.declareFlags(flags)

	flags.StringVar(
		&c.HomeBoot, "homeboot", "homedrv/homeboot",
		"init docker image",
	)
	flags.StringVar(
		&c.GitHubKeys, "github_keys", "",
		"add the ssh keys of a github user",
	)
	flags.StringVar(
		&c.UserKeys, "user_keys", "",
		"add the ssh keys of a homedrive user",
	)
}
