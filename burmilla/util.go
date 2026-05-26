package burmilla

import (
	"shanhu.io/std/errcode"
)

func execError(ret int, err error) error {
	if err != nil {
		return err
	}
	if ret != 0 {
		return errcode.Internalf("exit value: %d", ret)
	}
	return nil
}
