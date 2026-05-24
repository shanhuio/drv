package jarvis

import (
	"testing"

	"shanhu.io/homedrv/drv/drvapi"
)

func TestManifestFromRelease(t *testing.T) {
	var empty drvapi.Release
	m := manifestFromRelease(&empty)
	if len(m) != 0 {
		t.Errorf("got %v, want empty map", m)
	}

	m2 := manifestFromRelease(nil)
	if len(m2) != 0 {
		t.Errorf("got %v, want empty map", m)
	}
}
