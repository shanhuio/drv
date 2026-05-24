package executil

import (
	"shanhu.io/g/errcode"
)

// RetError wraps the return value and the error. If err is not nil, it
// return err. When err is nil, if ret is not 0, it returns an internal
// error.
func RetError(ret int, err error) error {
	if err != nil {
		return err
	}
	if ret != 0 {
		return errcode.Internalf("exit value: %d", ret)
	}
	return nil
}
