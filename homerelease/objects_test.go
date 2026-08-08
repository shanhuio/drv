package homerelease

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteObjects(t *testing.T) {
	dir := t.TempDir()

	// Two source files keyed by their (fake) checksum names.
	srcA := filepath.Join(dir, "a.src")
	srcB := filepath.Join(dir, "b.src")
	if err := os.WriteFile(srcA, []byte("content-a"), 0644); err != nil {
		t.Fatalf("write srcA: %v", err)
	}
	if err := os.WriteFile(srcB, []byte("content-b"), 0644); err != nil {
		t.Fatalf("write srcB: %v", err)
	}

	objects := map[string]string{
		"sha256:bbb": srcB,
		"sha256:aaa": srcA,
	}

	out := filepath.Join(dir, "objs.tar")
	if err := writeObjects(out, objects); err != nil {
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
		got = append(got, entry{name: h.Name, body: string(bs)})
	}

	// Entries are written sorted by checksum name.
	want := []entry{
		{name: "sha256:aaa", body: "content-a"},
		{name: "sha256:bbb", body: "content-b"},
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
