package jarvis

import (
	"shanhu.io/g/settings"
)

type identityStore struct {
	settings settings.Settings
	key      string
}

func newIdentityStore(s settings.Settings, k string) *identityStore {
	return &identityStore{settings: s, key: k}
}

func (s *identityStore) Load(v any) error {
	return s.settings.Get(s.key, v)
}

func (s *identityStore) Check() (bool, error) {
	return s.settings.Has(s.key)
}

func (s *identityStore) Save(v any) error {
	return s.settings.Set(s.key, v)
}
