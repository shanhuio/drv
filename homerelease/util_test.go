package homerelease

import (
	"path/filepath"
	"regexp"
	"testing"
)

func TestFilePath(t *testing.T) {
	for _, test := range []struct {
		base  string
		parts []string
		want  string
	}{
		{base: "out", parts: nil, want: "out"},
		{base: "out", parts: []string{"a", "b"}, want: filepath.Join("out", "a", "b")},
		{
			base:  "/base",
			parts: []string{"shanhu.io/drv/dockers", "nextcloud33.imgsum"},
			want: filepath.Join(
				"/base", "shanhu.io", "drv", "dockers", "nextcloud33.imgsum",
			),
		},
		// A slash-separated part is converted to OS separators.
		{base: "b", parts: []string{"x/y/z"}, want: filepath.Join("b", "x", "y", "z")},
	} {
		got := filePath(test.base, test.parts...)
		if got != test.want {
			t.Errorf(
				"filePath(%q, %q) = %q, want %q",
				test.base, test.parts, got, test.want,
			)
		}
	}
}

func TestMakeReleaseName(t *testing.T) {
	// A non-dev channel does not need creds and yields a deterministic shape:
	// "<typ>-<YYYYMMDD>-<6 hex chars>".
	name, err := MakeReleaseName("prod")
	if err != nil {
		t.Fatalf("MakeReleaseName: %v", err)
	}
	re := regexp.MustCompile(`^prod-[0-9]{8}-[0-9a-f]{6}$`)
	if !re.MatchString(name) {
		t.Errorf("MakeReleaseName(%q) = %q, want match %q", "prod", name, re)
	}
}
