package homeboot

import (
	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
	"shanhu.io/g/osutil"
	"shanhu.io/g/tarutil"
	drvcfg "shanhu.io/homedrv/drv/drvconfig"
)

// CoreMount is the mount point of jarvis volume.
const CoreMount = "/opt/jarvis/var"

// CoreConfig specifies how to start a core.
type CoreConfig struct {
	Drive *drvcfg.Config
	Image string
	Files *tarutil.Stream

	BindSysDock bool // bind system-docker.sock
}

// StartCore starts the core.homedrv container.
func StartCore(client *dock.Client, config *CoreConfig) (string, error) {
	naming := config.Drive.Naming
	image := config.Image
	if image == "" {
		return "", errcode.InvalidArgf("image missing")
	}
	name := drvcfg.Core(naming)
	labels := drvcfg.NewNameLabel("core")

	binds := []*dock.ContMount{{
		Type: dock.MountVolume,
		Host: name,
		Cont: CoreMount,
	}}

	bindSocks := []string{dock.Socket}
	if config.BindSysDock {
		bindSocks = append(bindSocks, systemDockSock)
	}
	for _, s := range bindSocks {
		ok, err := osutil.IsSock(s)
		if err != nil {
			return "", errcode.Annotatef(err, "check socket %q", s)
		}
		if !ok {
			return "", errcode.Annotatef(err, "socket %q", s)
		}
		binds = append(binds, &dock.ContMount{Host: s, Cont: s})
	}

	dockConfig := &dock.ContConfig{
		Name:        name,
		Network:     drvcfg.Network(naming),
		AutoRestart: true,
		Mounts:      binds,
		Labels:      labels,
	}
	// Note that core cannot bind ports, either TCP or UDP,
	// because updating the core needs to run a new one along
	// side with the old one.

	if _, err := dock.CreateVolumeIfNotExist(
		client, name, &dock.VolumeConfig{Labels: labels},
	); err != nil {
		return "", errcode.Annotate(err, "create volume for core")
	}

	cont, err := dock.CreateCont(client, image, dockConfig)
	if err != nil {
		return "", errcode.Annotate(err, "create core container")
	}

	if config.Files != nil {
		if err := dock.CopyInTarStream(
			cont, config.Files, CoreMount,
		); err != nil {
			cont.Drop()
			return "", errcode.Annotate(err, "copy in init files")
		}
	}

	if err := cont.Start(); err != nil {
		return "", errcode.Annotate(err, "start core container")
	}
	return cont.ID(), nil
}
