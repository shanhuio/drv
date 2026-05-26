package homerelease

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"path"

	"shanhu.io/g/httputil"
	"shanhu.io/std/errcode"
)

// Uploader is an object uploader.
type Uploader struct {
	Client *httputil.Client

	Prefix  string
	DataURL string
	APIURL  string
}

func (u *Uploader) key(k string) string {
	if u.Prefix == "" {
		return k
	}
	return path.Join(u.Prefix, k)
}

func (u *Uploader) exists(h string) (bool, error) {
	var found bool
	apiPath := path.Join(u.APIURL, "exists")
	if err := u.Client.Call(apiPath, u.key(h), &found); err != nil {
		return false, err
	}
	return found, nil
}

func shortKey(k string) string {
	const n = 19
	if len(k) > n {
		return k[:n]
	}
	return k
}

// Upload uploads an objects tarball.
func (u *Uploader) Upload(objsFile string) error {
	f, err := os.Open(objsFile)
	if err != nil {
		return errcode.Annotate(err, "open objects file")
	}
	defer f.Close()

	t := tar.NewReader(f)
	for {
		h, err := t.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return errcode.Annotate(err, "read tar")
		}
		k := h.Name

		exists, err := u.exists(k)
		if err != nil {
			return errcode.Annotatef(err, "check exists %q", k)
		}
		if exists {
			continue
		}

		log.Printf("uploading %q (%d bytes)", shortKey(k), h.Size)
		p := path.Join(u.DataURL, u.key(k))
		if err := u.Client.PutN(p, t, h.Size); err != nil {
			return errcode.Annotatef(err, "upload %q", k)
		}
	}

	return nil
}
