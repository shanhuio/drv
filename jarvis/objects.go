package jarvis

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
	"shanhu.io/g/hashutil"
	"shanhu.io/g/osutil"
)

type objects struct {
	dir string
}

func newObjects(dir string) (*objects, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, errcode.Annotate(err, "make objects dir")
	}
	return &objects{dir: dir}, nil
}

func (b *objects) writeFile(p string, r io.Reader) error {
	fp := filepath.Join(b.dir, p)
	f, err := os.Create(fp)
	if err != nil {
		return errcode.Annotate(err, "create file")
	}
	ok := false
	defer func() {
		f.Close()
		if !ok {
			os.Remove(fp)
		}
	}()

	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	if err := f.Sync(); err != nil {
		return errcode.Annotate(err, "sync to storage")
	}
	ok = true
	return nil
}

func (b *objects) Serve(c *aries.C) error {
	if c.User != rootUser {
		return errcode.Unauthorizedf("only root can use this")
	}

	p := c.Rel()
	if p == "" {
		return errcode.InvalidArgf("path is empty")
	}
	if strings.Contains(p, "/") {
		return errcode.InvalidArgf("path contains slash")
	}

	switch c.Req.Method {
	case http.MethodGet, http.MethodHead:
		http.ServeFile(c.Resp, c.Req, filepath.Join(b.dir, p))
		return nil
	case http.MethodPut:
		n := c.Req.ContentLength
		if n < 0 {
			return errcode.InvalidArgf("content length missing")
		}
		cr, err := hashutil.NewCheckReader(c.Req.Body, p, n)
		if err != nil {
			return errcode.Annotate(err, "check hash")
		}

		if err := b.writeFile(p, cr); err != nil {
			return errcode.Annotate(err, "save object")
		}
		return nil
	default:
		return errcode.InvalidArgf("unsupported method: %q", c.Req.Method)
	}
}

func (b *objects) open(p string) (*os.File, error) {
	return os.Open(filepath.Join(b.dir, p))
}

func (b *objects) apiExists(c *aries.C, p string) (bool, error) {
	return osutil.IsRegular(filepath.Join(b.dir, p))
}

func (b *objects) api() *aries.Router {
	r := aries.NewRouter()
	r.Call("exists", b.apiExists)
	return r
}
