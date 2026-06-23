package bosinit

import (
	"fmt"
	"net/url"
	"strings"

	"shanhu.io/g/httputil"
)

// FetchGitHubKeys fetches github ssh public keys of a user.
func FetchGitHubKeys(user string) ([]string, error) {
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
	for _, line := range strings.Split(keys, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines, nil
}
