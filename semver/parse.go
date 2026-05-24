package semver

import (
	"strconv"
	"strings"

	"shanhu.io/g/errcode"
)

// Major returns the major version number of a version string.
func Major(v string) (int, error) {
	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return 0, errcode.InvalidArgf("invalid version: %q", v)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, errcode.Annotatef(err, "parse major version: %q", v)
	}
	if major <= 0 {
		return 0, errcode.InvalidArgf("invalid major version in %q", v)
	}
	return major, nil
}
