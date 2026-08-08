package homerelease

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"shanhu.io/std/errcode"
	"shanhu.io/std/tarutil"
)

// writeObjects writes the objects into a tarball keyed by their checksums. It
// also writes a manifest.json entry (first, so a reader sees it before the
// objects) that maps each object's checksum to its source file path relative
// to baseDir, so uploads can be labeled with a human-readable name.
func writeObjects(p string, objects map[string]string, baseDir string) error {
	var sums []string
	for sum := range objects {
		sums = append(sums, sum)
	}
	sort.Strings(sums)

	names := make(map[string]string)
	for sum, file := range objects {
		name, err := filepath.Rel(baseDir, file)
		if err != nil {
			return errcode.Annotatef(err, "rel path for %q", file)
		}
		names[sum] = name
	}

	manifest, err := json.Marshal(&objectManifest{Names: names})
	if err != nil {
		return errcode.Annotate(err, "marshal manifest")
	}

	stream := tarutil.NewStream()
	stream.AddBytes(objectManifestName, tarutil.ModeMeta(0644), manifest)
	for _, sum := range sums {
		stream.AddFile(sum, tarutil.ModeMeta(0644), objects[sum])
	}

	f, err := os.Create(p)
	if err != nil {
		return errcode.Annotate(err, "create file")
	}
	defer f.Close()

	if _, err := stream.WriteTo(f); err != nil {
		return errcode.Annotate(err, "write tarball")
	}
	if err := f.Sync(); err != nil {
		return errcode.Annotate(err, "sync to disk")
	}
	return nil
}
