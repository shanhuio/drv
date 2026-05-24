package homeboot

import (
	"log"

	"shanhu.io/homedrv/drv/drvapi"
	drvcfg "shanhu.io/homedrv/drv/drvconfig"
	"shanhu.io/homedrv/drv/semver"
)

// DownloadConfig is the install config. This is the configuration
// for downloading and installing.
type DownloadConfig struct {
	Release *drvapi.Release
	Channel string
	Build   string

	Naming *drvcfg.Naming // Naming conventions.

	// Download the core only; only used in homeboot for bootstraping.
	CoreOnly bool

	// Only downloads the latest one from the ladder.
	LatestOnly bool

	// If set, ignore major versions that are lower than this.
	CurrentSemVersions map[string]string
}

func (c *DownloadConfig) currentMajor(app string) int {
	if c.CurrentSemVersions == nil {
		return 0
	}
	v, ok := c.CurrentSemVersions[app]
	if !ok {
		return 0
	}
	if v == "" {
		return 0
	}
	major, err := semver.Major(v)
	if err != nil {
		log.Printf("invalid sem version of %q: %q: %s", app, v, err)
		return 0
	}
	return major
}
