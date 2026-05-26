package homeboot

import (
	"log"
	"strings"

	drvcfg "shanhu.io/drv/drvconfig"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
)

func findNameWithSuffix(suf string, names ...string) (string, bool) {
	for _, name := range names {
		if strings.HasSuffix(name, suf) {
			return name, true
		}
	}
	return "", false
}

func cmdUninstall(args []string) error {
	flags := cmdFlags.New()
	dockerSock := flags.String("docker", "", "docker unix domain socket")
	keepNetwork := flags.Bool("keep_network", false, "keep network")
	keepVolumes := flags.Bool("keep_volumes", false, "keep volumes")
	keepImages := flags.Bool("keep_image", false, "keep images")
	flags.ParseArgs(args)

	d := docker.NewUnixClient(*dockerSock)

	const suffix = drvcfg.DefaultSuffix

	conts, err := docker.ListContsWithLabel(d, drvcfg.LabelName)
	if err != nil {
		return errcode.Annotate(err, "list containers")
	}
	for _, cont := range conts {
		if name, ok := findNameWithSuffix(suffix, cont.Names...); ok {
			log.Printf("remove container %q", name)
			c := docker.NewCont(d, cont.ID)
			if err := c.ForceRemove(); err != nil {
				return errcode.Annotatef(err, "remove container %q", name)
			}
		}
	}

	if !*keepVolumes {
		vols, err := docker.ListVolumesWithLabel(d, drvcfg.LabelName)
		if err != nil {
			return errcode.Annotate(err, "list volumes")
		}
		for _, vol := range vols {
			name := vol.Name
			if !strings.HasSuffix(name, suffix) {
				continue
			}
			log.Printf("remove volume %q", name)
			if err := docker.RemoveVolume(d, name); err != nil {
				return errcode.Annotatef(err, "remove volume %q", name)
			}
		}
	}

	if !*keepNetwork {
		const network = drvcfg.DefaultNetwork
		if has, err := docker.HasNetwork(d, network); err != nil {
			return errcode.Annotate(err, "check network")
		} else if has {
			log.Printf("remove network %q", network)
			if err := docker.RemoveNetwork(d, network); err != nil {
				return errcode.Annotate(err, "remove network")
			}
		}
	}

	if !*keepImages {
		images, err := docker.ListImages(d)
		if err != nil {
			return errcode.Annotate(err, "list images")
		}
		tagPrefix := drvcfg.DefaultRegistry + "/"
		removeOpt := &docker.RemoveImageOptions{}
		for _, img := range images {
			for _, tag := range img.RepoTags {
				if !strings.HasPrefix(tag, tagPrefix) {
					continue
				}
				log.Printf("remove image %q", tag)
				if err := docker.RemoveImage(d, tag, removeOpt); err != nil {
					return errcode.Annotatef(err, "remove image %q", tag)
				}
			}
		}
	}

	log.Println("homedrive uninstalled.")
	return nil
}
