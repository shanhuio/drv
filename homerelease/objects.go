package homerelease

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"shanhu.io/std/errcode"
	"shanhu.io/std/tarutil"
)

// writeObjects writes the objects into a tarball keyed by their checksums.
// objects maps each object's checksum to its source file path relative to
// baseDir. It also writes a manifest.json entry (first, so a reader sees it
// before the objects) carrying those relative paths, so uploads can be labeled
// with a human-readable name.
func writeObjects(p string, objects map[string]string, baseDir string) error {
	var sums []string
	for sum := range objects {
		sums = append(sums, sum)
	}
	// Order the objects by their file name for a stable, human-friendly order.
	sort.Slice(sums, func(i, j int) bool {
		return objects[sums[i]] < objects[sums[j]]
	})

	manifest, err := json.Marshal(&objectManifest{Names: objects})
	if err != nil {
		return errcode.Annotate(err, "marshal manifest")
	}

	stream := tarutil.NewStream()
	stream.AddBytes(objectManifestName, tarutil.ModeMeta(0644), manifest)
	for _, sum := range sums {
		file := filepath.Join(baseDir, objects[sum])
		stream.AddFile(sum, tarutil.ModeMeta(0644), file)
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
