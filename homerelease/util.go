package homerelease

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"shanhu.io/g/creds"
	"shanhu.io/g/errcode"
	"shanhu.io/g/rand"
)

// MakeReleaseName makes a new release name.
func MakeReleaseName(typ string) (string, error) {
	ch := typ
	if typ == "dev" {
		u, err := creds.CurrentUser()
		if err != nil {
			return "", errcode.Annotate(err, "get current user")
		}
		ch = "dev-" + u
	}

	date := time.Now().Format("20060102")
	return fmt.Sprintf("%s-%s-%s", ch, date, rand.HexBytes(3)), nil
}

func filePath(base string, parts ...string) string {
	p := path.Join(parts...)
	return filepath.Join(base, filepath.FromSlash(p))
}
