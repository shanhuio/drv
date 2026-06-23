package homeboot

import (
	"fmt"
	"net/url"
	"strings"

	"shanhu.io/drv/drvapi"
	"shanhu.io/g/httputil"
	"shanhu.io/std/errcode"
)

func fetchGitHubKeys(user string) ([]string, error) {
	c := &httputil.Client{
		Server: &url.URL{
			Scheme: "https",
			Host:   "github.com",
		},
	}

	keys, err := c.GetString(fmt.Sprintf("/%s.keys", user))
	if err != nil {
		return nil, err
	}

	var lines []string
	for line := range strings.SplitSeq(keys, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

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
		keys, err := fetchGitHubKeys(c.GitHubKeys)
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
