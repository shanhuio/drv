package homerelease

import (
	"archive/tar"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestWriteObjects(t *testing.T) {
	dir := t.TempDir()

	// Two source files keyed by their (fake) checksum names. srcA sits in a
	// subdirectory so the manifest name is a relative path, not a basename.
	subDir := filepath.Join(dir, "dockers")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("mkdir subdir: %v", err)
	}
	srcA := filepath.Join(subDir, "a.src")
	srcB := filepath.Join(dir, "b.src")
	if err := os.WriteFile(srcA, []byte("content-a"), 0644); err != nil {
		t.Fatalf("write srcA: %v", err)
	}
	if err := os.WriteFile(srcB, []byte("content-b"), 0644); err != nil {
		t.Fatalf("write srcB: %v", err)
	}

	objects := map[string]string{
		"fake:bbb": srcB,
		"fake:aaa": srcA,
	}

	out := filepath.Join(dir, "objs.tar")
	if err := writeObjects(out, objects, dir); err != nil {
		t.Fatalf("writeObjects: %v", err)
	}

	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("open out: %v", err)
	}
	defer f.Close()

	type entry struct {
		name string
		body string
	}
	var got []entry
	var manifest *objectManifest
	tr := tar.NewReader(f)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("read tar: %v", err)
		}
		bs, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("read entry %q: %v", h.Name, err)
		}
		if h.Name == objectManifestName {
			// The manifest must come before any object entry so a reader
			// sees it first.
			if len(got) != 0 {
				t.Errorf("manifest is not the first entry")
			}
			manifest = new(objectManifest)
			if err := json.Unmarshal(bs, manifest); err != nil {
				t.Fatalf("unmarshal manifest: %v", err)
			}
			continue
		}
		got = append(got, entry{name: h.Name, body: string(bs)})
	}

	if manifest == nil {
		t.Fatalf("no manifest entry found")
	}
	wantNames := map[string]string{
		"fake:aaa": filepath.Join("dockers", "a.src"),
		"fake:bbb": "b.src",
	}
	if !reflect.DeepEqual(manifest.Names, wantNames) {
		t.Errorf("manifest names = %+v, want %+v", manifest.Names, wantNames)
	}

	// Object entries are written sorted by checksum name.
	want := []entry{
		{name: "fake:aaa", body: "content-a"},
		{name: "fake:bbb", body: "content-b"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d: %+v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("entry %d = %+v, want %+v", i, got[i], w)
		}
	}
}
