package sniproxy

import (
	"errors"
	"testing"
)

func TestFirstErr(t *testing.T) {
	var e firstErr

	// Nothing set yet.
	if got := e.get(); got != nil {
		t.Fatalf("get on empty: got %v, want nil", got)
	}

	// Setting nil keeps it empty.
	e.set(nil)
	if got := e.get(); got != nil {
		t.Fatalf("get after set(nil): got %v, want nil", got)
	}

	// The first non-nil error sticks.
	first := errors.New("first")
	e.set(first)
	if got := e.get(); got != first {
		t.Fatalf("get after first set: got %v, want %v", got, first)
	}

	// Later errors, including nil, do not overwrite the first.
	e.set(nil)
	e.set(errors.New("second"))
	if got := e.get(); got != first {
		t.Fatalf("get after later sets: got %v, want %v", got, first)
	}
}
