package jarvis

import (
	"sort"

	"shanhu.io/g/errcode"
	"shanhu.io/g/pisces"
	"shanhu.io/homedrv/drv/homeapp"
)

type appDomains struct {
	t *pisces.KV
}

func newAppDomains(b *pisces.Tables) *appDomains {
	return &appDomains{t: b.NewKV("app_domains")}
}

func (b *appDomains) Set(m *homeapp.DomainMap) error {
	if len(m.Map) == 0 {
		return b.Clear(m.App)
	}
	return b.t.Replace(m.App, m)
}

func (b *appDomains) Clear(app string) error {
	if err := b.t.Remove(app); err != nil {
		if errcode.IsNotFound(err) {
			return nil
		}
		return err
	}
	return nil
}

func (b *appDomains) list() ([]*homeapp.DomainMap, error) {
	var maps []*homeapp.DomainMap
	it := &pisces.Iter{
		Make: func() interface{} { return new(homeapp.DomainMap) },
		Do: func(_ string, v interface{}) error {
			maps = append(maps, v.(*homeapp.DomainMap))
			return nil
		},
	}
	if err := b.t.Walk(it); err != nil {
		return nil, err
	}
	sort.Slice(maps, func(i, j int) bool {
		return maps[i].App < maps[j].App
	})
	return maps, nil
}
