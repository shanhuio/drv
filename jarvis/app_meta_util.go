package jarvis

import (
	"shanhu.io/drv/drvapi"
)

func makeManifest(metas []*drvapi.AppMeta) map[string]*drvapi.AppMeta {
	m := make(map[string]*drvapi.AppMeta)
	for _, meta := range metas {
		m[meta.Name] = meta
	}
	return m
}

type appQuerier interface {
	// Returns the meta info for an app. Returns NotFound error if app not
	// found.
	meta(name string) (*drvapi.AppMeta, error)
}
