package homerelease

import (
	"os"
	"path/filepath"
	"testing"

	"shanhu.io/drv/drvapi"
)

// writeImgsum writes a lets-style <name>.imgsum file with the given origin and
// returns its path.
func writeImgsum(t *testing.T, dir, name, origin string) string {
	t.Helper()
	p := filepath.Join(dir, name+".imgsum")
	// Only the Origin field matters to readImageStep; mimic the real file,
	// which also carries Repo/Tag/ID.
	content := `{"Repo":"cr.shanhu.io/drv/` + name + `","Tag":"latest",` +
		`"ID":"sha256:deadbeef","Origin":"` + origin + `"}`
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write imgsum: %v", err)
	}
	return p
}

func TestReadImageStep(t *testing.T) {
	dir := t.TempDir()
	img := &dockerImage{
		id:  "sha256:abc123",
		sum: "sha256:sumsumsum",
	}

	for _, test := range []struct {
		name       string
		origin     string
		wantMajor  int
		wantVer    string
		wantSource string
	}{
		{
			name:       "nextcloud33",
			origin:     "nextcloud:33.0.7",
			wantMajor:  33,
			wantVer:    "33.0.7",
			wantSource: "nextcloud:33.0.7",
		},
		{
			// A fully-qualified registry path: version comes from the tag,
			// not the path.
			name:       "registry-qualified",
			origin:     "cr.example.com/library/nextcloud:32.0.13",
			wantMajor:  32,
			wantVer:    "32.0.13",
			wantSource: "cr.example.com/library/nextcloud:32.0.13",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			p := writeImgsum(t, dir, test.name, test.origin)
			step, err := readImageStep(p, img)
			if err != nil {
				t.Fatalf("readImageStep: %v", err)
			}
			want := &drvapi.StepVersion{
				Major:    test.wantMajor,
				Version:  test.wantVer,
				Source:   test.wantSource,
				Image:    img.id,
				ImageSum: img.sum,
			}
			if *step != *want {
				t.Errorf("step = %+v, want %+v", *step, *want)
			}
		})
	}
}

func TestReadImageStepErrors(t *testing.T) {
	dir := t.TempDir()
	img := &dockerImage{id: "sha256:abc", sum: "sha256:def"}

	t.Run("missing file", func(t *testing.T) {
		p := filepath.Join(dir, "nope.imgsum")
		if _, err := readImageStep(p, img); err == nil {
			t.Fatalf("got nil error, want read error")
		}
	})

	t.Run("empty origin", func(t *testing.T) {
		p := writeImgsum(t, dir, "no-origin", "")
		if _, err := readImageStep(p, img); err == nil {
			t.Fatalf("got nil error, want origin-missing error")
		}
	})

	t.Run("origin without tag", func(t *testing.T) {
		p := writeImgsum(t, dir, "no-tag", "nextcloud")
		if _, err := readImageStep(p, img); err == nil {
			t.Fatalf("got nil error, want tag-parse error")
		}
	})

	t.Run("non-numeric tag", func(t *testing.T) {
		p := writeImgsum(t, dir, "latest-tag", "nextcloud:latest")
		if _, err := readImageStep(p, img); err == nil {
			t.Fatalf("got nil error, want major-parse error")
		}
	})
}
