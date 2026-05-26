package homeboot

import (
	"shanhu.io/g/osutil"
	"shanhu.io/std/errcode"
)

const systemDockSock = "/var/run/system-docker.sock"

// CheckSystemDock checks if system docker exists. It returns NotFound error if
// not.
func CheckSystemDock() error {
	ok, err := osutil.IsSock(systemDockSock)
	if err != nil {
		return err
	}
	if !ok {
		return errcode.NotFoundf("system docker not found")
	}
	return nil
}
