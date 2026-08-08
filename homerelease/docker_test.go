package homerelease

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// makeDockerTgz builds a gzipped tar that looks like a "docker save" output:
// a manifest.json listing the given config references, plus any extra files.
func makeDockerTgz(t *testing.T, configs []string, extra map[string]string) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)

	writeFile := func(name string, data []byte) {
		hdr := &tar.Header{
			Typeflag: tar.TypeReg,
			Name:     name,
			Mode:     0644,
			Size:     int64(len(data)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("write header %q: %v", name, err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatalf("write %q: %v", name, err)
		}
	}

	for name, content := range extra {
		writeFile(name, []byte(content))
	}

	if configs != nil {
		var m []*dockerManifest
		for _, c := range configs {
			m = append(m, &dockerManifest{Config: c})
		}
		bs, err := json.Marshal(m)
		if err != nil {
			t.Fatalf("marshal manifest: %v", err)
		}
		writeFile("manifest.json", bs)
	}

	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}

func TestExtractDockerManifest(t *testing.T) {
	tgz := makeDockerTgz(t, []string{"abc123.json"}, map[string]string{
		"some-layer/layer.tar": "junk",
	})
	m, err := extractDockerManifest(bytes.NewReader(tgz))
	if err != nil {
		t.Fatalf("extractDockerManifest: %v", err)
	}
	if len(m) != 1 {
		t.Fatalf("got %d manifests, want 1", len(m))
	}
	if m[0].Config != "abc123.json" {
		t.Errorf("Config = %q, want %q", m[0].Config, "abc123.json")
	}
}

func TestExtractDockerManifestNotFound(t *testing.T) {
	// A valid gzip tar without a manifest.json.
	tgz := makeDockerTgz(t, nil, map[string]string{"foo": "bar"})
	if _, err := extractDockerManifest(bytes.NewReader(tgz)); err == nil {
		t.Fatalf("got nil error, want manifest-not-found")
	}
}

func TestExtractDockerManifestBadGzip(t *testing.T) {
	if _, err := extractDockerManifest(bytes.NewReader([]byte("not gzip"))); err == nil {
		t.Fatalf("got nil error, want gzip error")
	}
}

func TestSumDockerTgz(t *testing.T) {
	for _, test := range []struct {
		name   string
		config string
		wantID string
	}{
		// Plain hex id gets a sha256: prefix.
		{name: "plain", config: "abc123.json", wantID: "sha256:abc123"},
		// OCI-style blobs path is rewritten to sha256:<hex>.
		{
			name:   "blobs",
			config: "blobs/sha256/deadbeef",
			wantID: "sha256:deadbeef",
		},
		// An already-qualified id is left as-is.
		{
			name:   "qualified",
			config: "sha256:cafef00d",
			wantID: "sha256:cafef00d",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			tgz := makeDockerTgz(t, []string{test.config}, nil)
			dir := t.TempDir()
			p := filepath.Join(dir, "image.tar.gz")
			if err := os.WriteFile(p, tgz, 0644); err != nil {
				t.Fatalf("write tgz: %v", err)
			}

			img, err := sumDockerTgz(p)
			if err != nil {
				t.Fatalf("sumDockerTgz: %v", err)
			}
			if img.id != test.wantID {
				t.Errorf("id = %q, want %q", img.id, test.wantID)
			}

			sum := sha256.Sum256(tgz)
			wantSum := "sha256:" + hex.EncodeToString(sum[:])
			if img.sum != wantSum {
				t.Errorf("sum = %q, want %q", img.sum, wantSum)
			}
		})
	}
}

func TestSumDockerTgzErrors(t *testing.T) {
	dir := t.TempDir()

	t.Run("missing file", func(t *testing.T) {
		if _, err := sumDockerTgz(filepath.Join(dir, "nope.tar.gz")); err == nil {
			t.Fatalf("got nil error, want open error")
		}
	})

	t.Run("multiple images", func(t *testing.T) {
		tgz := makeDockerTgz(t, []string{"a.json", "b.json"}, nil)
		p := filepath.Join(dir, "multi.tar.gz")
		if err := os.WriteFile(p, tgz, 0644); err != nil {
			t.Fatalf("write tgz: %v", err)
		}
		if _, err := sumDockerTgz(p); err == nil {
			t.Fatalf("got nil error, want multi-image error")
		}
	})

	t.Run("empty id", func(t *testing.T) {
		tgz := makeDockerTgz(t, []string{".json"}, nil)
		p := filepath.Join(dir, "empty.tar.gz")
		if err := os.WriteFile(p, tgz, 0644); err != nil {
			t.Fatalf("write tgz: %v", err)
		}
		if _, err := sumDockerTgz(p); err == nil {
			t.Fatalf("got nil error, want empty-id error")
		}
	})
}
