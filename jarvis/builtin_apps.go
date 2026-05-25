package jarvis

import (
	"shanhu.io/g/errcode"
	"shanhu.io/drv/homeapp"
	"shanhu.io/drv/homeapp/nextcloud"
	"shanhu.io/drv/homeapp/postgres"
	"shanhu.io/drv/homeapp/redis"
)

type builtInApps struct {
	stubs map[string]*appStub
}

func newBuiltInApps(c homeapp.Core) *builtInApps {
	m := make(map[string]*appStub)
	for _, a := range []struct {
		name string
		app  homeapp.App
	}{
		{name: "redis", app: redis.New(c)},
		{name: "postgres", app: postgres.New(c)},
		{name: "ncfront", app: nextcloud.NewFront(c)},
		{name: "nextcloud", app: nextcloud.New(c)},
	} {
		m[a.name] = &appStub{App: a.app}
	}

	return &builtInApps{stubs: m}
}

func (b *builtInApps) makeStub(name string) (*appStub, error) {
	a, ok := b.stubs[name]
	if ok {
		return a, nil
	}
	return nil, errcode.NotFoundf("app %q not found", name)
}
