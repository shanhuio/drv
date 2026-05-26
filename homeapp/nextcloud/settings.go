package nextcloud

import (
	"fmt"
	"sort"

	"shanhu.io/drv/homeapp"
	"shanhu.io/drv/homeapp/apputil"
	"shanhu.io/drv/homeapp/redis"
	"shanhu.io/g/settings"
	"shanhu.io/std/errcode"
)

// Domains reads the nextcloud domains from the settings.
func Domains(s settings.Settings) ([]string, error) {
	var domains []string
	if err := s.Get(KeyDomains, &domains); err == nil {
		return domains, nil
	} else if !errcode.IsNotFound(err) {
		return nil, err
	}
	// Domain list not found.

	set := func(domains []string) ([]string, error) {
		if err := s.Set(KeyDomains, domains); err != nil {
			return nil, errcode.Annotate(err, "set nextcloud domains")
		}
		return domains, nil
	}

	domain, err := settings.String(s, KeyDomain)
	if err == nil {
		return set([]string{domain})
	}
	if !errcode.IsNotFound(err) {
		return nil, errcode.Annotate(err, "read nextcloud domain")
	}
	// Single domain not found.

	main, err := settings.String(s, homeapp.KeyMainDomain)
	if err != nil {
		return nil, errcode.Annotate(err, "cannot determine domain")
	}
	return set([]string{fmt.Sprintf("nextcloud.%s", main)})
}

func loadConfig(c homeapp.Core) (*config, error) {
	s := c.Settings()

	// TODO(h8liu): reading redis password should to go redis?
	redisPass, err := settings.String(s, redis.KeyPass)
	if err != nil {
		return nil, errcode.Annotate(err, "read redis password")
	}

	adminPass, err := apputil.ReadPasswordOrSetRandom(s, KeyAdminPass)
	if err != nil {
		return nil, errcode.Annotate(err, "read init password")
	}
	dbPass, err := apputil.ReadPasswordOrSetRandom(s, KeyDBPass)
	if err != nil {
		return nil, errcode.Annotate(err, "read db password")
	}
	domains, err := Domains(s)
	if err != nil {
		return nil, errcode.Annotate(err, "load domains")
	}

	dataMount, err := settings.String(s, KeyDataMount)
	if err != nil {
		if errcode.IsNotFound(err) {
			dataMount = ""
		} else {
			return nil, errcode.Annotate(err, "read nextcloud data mount")
		}
	}

	var extraMountMap map[string]string
	if err := s.Get(KeyExtraMounts, &extraMountMap); err != nil {
		if errcode.IsNotFound(err) {
			extraMountMap = nil
		} else {
			return nil, errcode.Annotate(err, "read nextcloud extra mounts")
		}
	}

	var extraMounts []*extraMount
	if len(extraMountMap) > 0 {
		for k, v := range extraMountMap {
			extraMounts = append(extraMounts, &extraMount{
				host:      k,
				container: v,
			})
		}
		sort.Slice(extraMounts, func(i, j int) bool {
			return extraMounts[i].host < extraMounts[j].host
		})
	}

	return &config{
		domains:       domains,
		dbPassword:    dbPass,
		adminPassword: adminPass,
		redisPassword: redisPass,
		dataMount:     dataMount,
		extraMounts:   extraMounts,
	}, nil
}
