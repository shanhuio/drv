package jarvis

import (
	"shanhu.io/g/jsonx"
	"shanhu.io/g/osutil"
	drvcfg "shanhu.io/drv/drvconfig"
)

func readConfig(h *osutil.Home) (*drvcfg.Config, error) {
	f := h.Var("config.jsonx")
	c := new(drvcfg.Config)
	if err := jsonx.ReadFile(f, c); err != nil {
		return nil, err
	}
	return c, nil
}
