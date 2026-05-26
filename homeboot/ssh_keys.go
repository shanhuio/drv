package homeboot

import (
	"net/url"

	"shanhu.io/drv/drvapi"
	"shanhu.io/g/bosinit"
	"shanhu.io/g/httputil"
	"shanhu.io/std/errcode"
)

func fetchUserKeys(user string) ([]string, error) {
	c := &httputil.Client{
		Server: &url.URL{
			Scheme: "https",
			Host:   "www.homedrive.io",
		},
	}
	resp := new(drvapi.UserSSHKeyLines)
	if err := c.Call("/pubapi/user/sshkeys", user, resp); err != nil {
		return nil, err
	}
	return resp.Keys, nil
}

// FetchSSHKeys fetches the SSH keys specified by the config.
func FetchSSHKeys(c *InitConfig) ([]string, error) {
	var lines []string
	if c.GitHubKeys != "" {
		keys, err := bosinit.FetchGitHubKeys(c.GitHubKeys)
		if err != nil {
			return nil, errcode.Annotate(err, "fetch github keys")
		}
		lines = append(lines, keys...)
	}

	if c.UserKeys != "" {
		keys, err := fetchUserKeys(c.UserKeys)
		if err != nil {
			return nil, errcode.Annotatef(
				err, "fetch ssh keys of %q", c.UserKeys,
			)
		}
		lines = append(lines, keys...)
	}

	return lines, nil
}
