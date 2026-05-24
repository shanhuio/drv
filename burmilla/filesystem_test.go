package burmilla

import (
	"testing"
)

func TestMkdirCmd(t *testing.T) {
	// Sanity check.
	if _, err := mkdirCmd("/h/r/.ssh", "rancher"); err != nil {
		t.Errorf("mkdirCmd failed %s", err)
	}
}
