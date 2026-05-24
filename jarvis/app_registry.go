package jarvis

import (
	"shanhu.io/g/errcode"
	"shanhu.io/homedrv/drv/drvapi"
)

type appRegistry struct {
	manifest map[string]*drvapi.AppMeta
}

func manifestFromRelease(rel *drvapi.Release) map[string]*drvapi.AppMeta {
	if rel == nil {
		return make(map[string]*drvapi.AppMeta)
	}
	metas := rel.Apps
	if metas == nil {
		if rel.Artifacts == nil {
			return make(map[string]*drvapi.AppMeta)
		}
		metas = drvapi.LegacyAppsFromArtifacts(rel.Artifacts)
	}
	return makeManifest(metas)
}

func newAppRegistry(rel *drvapi.Release) *appRegistry {
	manifest := manifestFromRelease(rel)
	return &appRegistry{
		manifest: manifest,
	}
}

func (r *appRegistry) meta(name string) (*drvapi.AppMeta, error) {
	meta, found := r.manifest[name]
	if !found {
		return nil, errcode.NotFoundf("app meta not found for %q", name)
	}
	return meta, nil
}

func (r *appRegistry) setRelease(rel *drvapi.Release) {
	r.manifest = manifestFromRelease(rel)
}
