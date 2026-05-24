package apputil

import (
	"testing"

	"shanhu.io/g/pisces"
	"shanhu.io/g/settings"
)

func TestReadPassword(t *testing.T) {
	tables := pisces.NewTables(nil)
	s := settings.NewTable(tables)

	const key = "password"
	pwd, err := ReadPasswordOrSetRandom(s, key)
	if err != nil {
		t.Fatal("set password: ", err)
	}

	again, err := ReadPasswordOrSetRandom(s, key)
	if err != nil {
		t.Fatal("read password: ", err)
	}
	if pwd != again {
		t.Errorf("password got %q, want/was %q", again, pwd)
	}
}
