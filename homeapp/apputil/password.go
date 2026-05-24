package apputil

import (
	"shanhu.io/g/errcode"
	"shanhu.io/g/rand"
	"shanhu.io/g/settings"
)

func randPassword() string {
	return rand.Letters(16)
}

// ReadPasswordOrSetRandom reads a string password or set a random one.
func ReadPasswordOrSetRandom(
	s settings.Settings, k string,
) (string, error) {
	pwd, err := settings.String(s, k)
	if err != nil {
		if errcode.IsNotFound(err) {
			pwd := randPassword()
			if err := s.Set(k, pwd); err != nil {
				return "", errcode.Annotate(err, "set password")
			}
			return pwd, nil
		}
		return "", errcode.Annotate(err, "read password")
	}
	return pwd, nil
}
