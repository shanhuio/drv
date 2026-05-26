package jarvis

import (
	"shanhu.io/g/aries"
	"shanhu.io/std/errcode"
)

func parsePostForm(c *aries.C) error {
	if c.Req.Method != "POST" {
		return errcode.InvalidArgf("request must be post")
	}
	if err := c.Req.ParseForm(); err != nil {
		return errcode.InvalidArgf("error parsing form: %v", err)
	}
	return nil
}
