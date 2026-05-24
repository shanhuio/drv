package jarvis

import (
	"strings"

	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
)

func signInRedirect(c *aries.C) error {
	// TODO(h8liu): add sign-in redirect URL.
	c.Redirect("/")
	return nil
}

func serveDashboard(s *server, c *aries.C) error {
	if c.Req.Method != "GET" {
		return errcode.InvalidArgf("request must be get")
	}

	aries.NeverCache(c)
	if c.User == "" {
		signInRedirect(c)
		return nil
	}

	d, err := newDashboardData(s, c, &DashboardDataRequest{
		Path: strings.TrimPrefix(c.Path, "/"),
	})
	if err != nil {
		return err
	}
	dat := struct{ Data *DashboardData }{Data: d}
	return s.tmpls.Serve(c, "dashboard.html", &dat)
}
