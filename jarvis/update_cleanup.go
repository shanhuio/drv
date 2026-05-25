package jarvis

import (
	"log"
	"strings"

	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
	"shanhu.io/drv/drvapi"
)

func releaseImagesToKeep(r *drvapi.Release) map[string]bool {
	m := make(map[string]bool)
	arts := r.Artifacts
	if arts != nil {
		for _, img := range []string{
			arts.Jarvis,
			arts.Doorway,
			arts.Toolbox,
			arts.NCFront,
			arts.Nextcloud,
			arts.Redis,
			arts.Postgres,
		} {
			if img == "" {
				continue
			}
			if !strings.Contains(img, ":") {
				img = "sha256:" + img
			}
			m[img] = true
		}
	}

	for _, app := range r.Apps {
		img := app.Image
		if !strings.Contains(img, ":") {
			img = "sha256:" + img
		}
		m[img] = true
	}
	return m
}

func looksLikeHomeDriveImage(repoTag string) bool {
	for _, prefix := range []string{
		"cr.shanhu.io/",
		"registry.digitalocean.com/shanhu/",
		"cr.homedrive.io/",
		"nextcloud:",
		"postgres:",
		"redis:",
		"ncfront:",
		"core:",
	} {
		if strings.HasPrefix(repoTag, prefix) {
			return true
		}
	}
	return false
}

func updateCleanUp(d *drive, r *drvapi.Release) error {
	keep := releaseImagesToKeep(r)

	images, err := dock.ListImages(d.dock)
	if err != nil {
		return errcode.Annotate(err, "list images")
	}
	removeOpts := &dock.RemoveImageOptions{NoPrune: true}
	for _, img := range images {
		if _, found := keep[img.ID]; found {
			continue
		}
		for _, t := range img.RepoTags {
			if strings.HasPrefix(t, "cr.homedrive.io/empty:") {
				continue
			}
			if strings.HasPrefix(t, "empty:") {
				continue
			}
			if looksLikeHomeDriveImage(t) {
				if err := dock.RemoveImage(
					d.dock, t, removeOpts,
				); err != nil {
					log.Printf("WARNING: untag %q: %s", t, err)
				}
			}
		}
	}

	opt := &dock.PruneImagesOption{}
	if err := dock.PruneImages(d.dock, opt); err != nil {
		return errcode.Annotate(err, "prune untagged docker images")
	}
	return nil
}
