package homeboot

import (
	"shanhu.io/std/errcode"
	"shanhu.io/std/jsonx"
)

func cmdInstall(args []string) error {
	config := newBootConfig()

	flags := cmdFlags.New()
	config.declareFlags(flags)
	configFile := flags.String("config_file", "", "config file")
	flags.ParseArgs(args)

	if *configFile != "" {
		config = newBootConfig() // clear the flag defaults
		if err := jsonx.ReadFile(*configFile, config); err != nil {
			return errcode.Annotate(err, "read boot config")
		}
	}

	config.fixLegacyNaming()
	b := newBoot(config)
	return b.run()
}
