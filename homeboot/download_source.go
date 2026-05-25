package homeboot

import (
	"io"
	"path"

	"shanhu.io/drv/drvapi"
	"shanhu.io/g/errcode"
	"shanhu.io/g/httputil"
)

// FetchChannelRelease fetch the release from a particular channel.
func FetchChannelRelease(c *httputil.Client, ch string) (
	*drvapi.Release, error,
) {
	r := new(drvapi.Release)
	const p = "/pubapi/release/channel"
	if err := c.Call(p, ch, r); err != nil {
		return nil, errcode.Annotate(err, "fetch channel")
	}
	return r, nil
}

// FetchBuildRelease fetches the a particular build.
func FetchBuildRelease(c *httputil.Client, b string) (*drvapi.Release, error) {
	r := new(drvapi.Release)
	const p = "/pubapi/release/get"
	if err := c.Call(p, b, r); err != nil {
		return nil, errcode.Annotate(err, "fetch release")
	}
	return r, nil
}

// DownloadSource is a source for downloading a release.
type DownloadSource struct {
	// Build gets the release by name.
	Build func(b string) (*drvapi.Release, error)

	// Channel gets the release of a channel.
	Channel func(ch string) (*drvapi.Release, error)

	// OpenObject opens an object by name.
	OpenObject func(name string) (io.ReadCloser, error)

	// OpenDocker is the legacy way to download a docker image.
	OpenDocker func(name, hash string) (io.ReadCloser, error)
}

// OfficialDownloadSource creates a downloader downloading from
// HomeDrive official website.
func OfficialDownloadSource(c *httputil.Client) *DownloadSource {
	return &DownloadSource{
		Build: func(b string) (*drvapi.Release, error) {
			return FetchBuildRelease(c, b)
		},
		Channel: func(ch string) (*drvapi.Release, error) {
			return FetchChannelRelease(c, ch)
		},
		OpenObject: func(name string) (io.ReadCloser, error) {
			p := path.Join("/dl/obj", name)
			req, err := c.Get(p)
			if err != nil {
				return nil, err
			}
			return req.Body, nil
		},
		OpenDocker: func(name, hash string) (io.ReadCloser, error) {
			p := path.Join("/dl/docker", name, hash+".tar.gz")
			req, err := c.Get(p)
			if err != nil {
				return nil, err
			}
			return req.Body, nil
		},
	}
}
