package jarvis

import (
	"strings"

	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
)

// DashboardSSHKeysData contains data for initializing SSH Keys dashboard page.
type DashboardSSHKeysData struct {
	Keys     string
	Disabled bool
}

func newDashboardSSHKeysData(s *server, _ *aries.C) (
	*DashboardSSHKeysData, error,
) {
	if !s.drive.hasSys() {
		return &DashboardSSHKeysData{Disabled: true}, nil
	}

	keys, err := s.sshKeys.list()
	if err != nil {
		return nil, errcode.Annotate(err, "get ssh keys")
	}

	dat := new(DashboardSSHKeysData)
	if len(keys) > 0 {
		dat.Keys = strings.Join(keys, "\n") + "\n"
	}

	return dat, nil
}
