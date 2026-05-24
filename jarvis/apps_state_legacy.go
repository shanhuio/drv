package jarvis

import (
	"shanhu.io/g/errcode"
	"shanhu.io/homedrv/drv/drvapi"
	"shanhu.io/homedrv/drv/homeapp/nextcloud"
	"shanhu.io/homedrv/drv/homeapp/postgres"
	"shanhu.io/homedrv/drv/homeapp/redis"
)

func appsStateForLegacyUpgrade(reg *appRegistry) (
	*appsState, error,
) {
	state := &appsState{
		Metas:    make(map[string]*drvapi.AppMeta),
		Anchored: map[string]bool{nextcloud.Name: true},
	}
	for _, name := range []string{
		redis.Name,
		postgres.Name,
		nextcloud.Name,
		nextcloud.NameFront,
	} {
		meta, err := reg.meta(name)
		if err != nil {
			return nil, err
		}
		state.Metas[name] = meta
	}
	return state, nil
}

func maybeSetAppsStateFromLegacy(
	s *appsStateSettings, reg *appRegistry,
) error {
	hasState, err := s.has()
	if err != nil {
		return errcode.Annotate(err, "check apps state")
	}
	if hasState {
		return nil
	}

	// no state yet, and we have nextcloud, so this is an upgrade
	// from legacy system.
	state, err := appsStateForLegacyUpgrade(reg)
	if err != nil {
		return errcode.Annotate(err, "build apps state")
	}
	return s.save(state)
}
