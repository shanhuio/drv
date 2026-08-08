package homerelease

import (
	"archive/tar"
	"encoding/json"
	"io"
	"log"
	"os"
	"path"

	"shanhu.io/g/httputil"
	"shanhu.io/std/errcode"
)

// objectManifestName is the name of the manifest entry in an objects tarball.
// It maps object checksums to human-readable image names and is not itself an
// object to upload.
const objectManifestName = "manifest.json"

// objectManifest describes the objects in an objects tarball.
type objectManifest struct {
	// Names maps an object's checksum key to a human-readable image name.
	Names map[string]string `json:",omitempty"`
}

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

	var names map[string]string
	first := true

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

		if first {
			first = false
			if k != objectManifestName {
				return errcode.Internalf(
					"manifest %q must be the first file, got %q",
					objectManifestName, k,
				)
			}
			m := new(objectManifest)
			if err := json.NewDecoder(t).Decode(m); err != nil {
				return errcode.Annotate(err, "read manifest")
			}
			names = m.Names
			continue
		}

		if k == objectManifestName {
			return errcode.Internalf("duplicate manifest entry")
		}

		name := names[k]
		if name == "" {
			return errcode.Internalf("object %q has no name in manifest", k)
		}

		exists, err := u.exists(k)
		if err != nil {
			return errcode.Annotatef(err, "check exists %q", k)
		}
		if exists {
			continue
		}

		log.Printf(
			"uploading %s (%s, %d bytes)", name, shortKey(k), h.Size,
		)
		p := path.Join(u.DataURL, u.key(k))
		if err := u.Client.PutN(p, t, h.Size); err != nil {
			return errcode.Annotatef(err, "upload %q", k)
		}
	}

	if first {
		return errcode.Internalf("no manifest found in objects tarball")
	}

	return nil
}
