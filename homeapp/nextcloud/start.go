package nextcloud

import (
	"strings"

	drvcfg "shanhu.io/drv/drvconfig"
	"shanhu.io/drv/homeapp"
	"shanhu.io/drv/homeapp/postgres"
	"shanhu.io/drv/homeapp/redis"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
)

type extraMount struct {
	host      string
	container string
}

type config struct {
	domains       []string
	dbPassword    string
	adminPassword string
	redisPassword string
	dataMount     string
	extraMounts   []*extraMount
}

func networkCIDRs(c homeapp.Core) ([]string, error) {
	network := homeapp.Network(c)
	info, err := docker.InspectNetwork(c.Docker(), network)
	if err != nil {
		return nil, err
	}
	if info.IPAM == nil {
		return nil, nil
	}
	var cidrs []string
	for _, c := range info.IPAM.Config {
		cidrs = append(cidrs, c.Subnet)
	}
	return cidrs, nil
}

func createCont(
	c homeapp.Core, image string, config *config,
) (*docker.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}
	labels := drvcfg.NewNameLabel(Name)
	volName := homeapp.Vol(c, Name)

	contConfig := &docker.ContConfig{
		Name:          homeapp.Cont(c, Name),
		Network:       homeapp.Network(c),
		AutoRestart:   true,
		JSONLogConfig: docker.LimitedJSONLog(),
		Labels:        labels,
	}

	cidrs, err := networkCIDRs(c)
	if err != nil {
		return nil, errcode.Annotate(err, "list network CIDRs")
	}

	contConfig.Mounts = append(contConfig.Mounts, &docker.ContMount{
		Type: docker.MountVolume,
		Host: volName,
		Cont: "/var/www/html",
	})
	if config.dataMount != "" {
		contConfig.Mounts = append(contConfig.Mounts, &docker.ContMount{
			Type: docker.MountBind,
			Host: config.dataMount,
			Cont: "/var/www/html/data",
		})
	}
	for _, extra := range config.extraMounts {
		contConfig.Mounts = append(contConfig.Mounts, &docker.ContMount{
			Type: docker.MountBind,
			Host: extra.host,
			Cont: extra.container,
		})
	}
	contConfig.Env = map[string]string{
		"POSTGRES_HOST":       homeapp.Cont(c, postgres.Name),
		"POSTGRES_DB":         "nextcloud",
		"POSTGRES_USER":       "nextcloud",
		"POSTGRES_PASSWORD":   config.dbPassword,
		"REDIS_HOST":          homeapp.Cont(c, redis.Name),
		"REDIS_HOST_PASSWORD": config.redisPassword,

		"NEXTCLOUD_ADMIN_USER":     "admin",
		"NEXTCLOUD_ADMIN_PASSWORD": config.adminPassword,

		"PHP_MEMORY_LIMIT": "2G",
	}
	if len(config.domains) > 0 {
		domains := strings.Join(config.domains, " ")
		contConfig.Env["NEXTCLOUD_TRUSTED_DOMAINS"] = domains
	}
	if len(cidrs) > 0 {
		proxies := strings.Join(cidrs, " ")
		contConfig.Env["TRUSTED_PROXIES"] = proxies
	}

	d := c.Docker()
	if _, err := docker.CreateVolumeIfNotExist(
		d, volName, &docker.VolumeConfig{Labels: labels},
	); err != nil {
		return nil, errcode.Annotate(err, "create volume")
	}
	return docker.CreateCont(d, image, contConfig)
}

func start(
	c homeapp.Core, image string, config *config,
) error {
	cont, err := createCont(c, image, config)
	if err != nil {
		return errcode.Annotate(err, "create nextcloud")
	}
	return cont.Start()
}
