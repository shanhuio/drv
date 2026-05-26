package strutil

import (
	"testing"

	"reflect"
)

func TestMakeSet(t *testing.T) {
	for _, test := range []struct {
		list []string
		want map[string]bool
	}{
		{nil, map[string]bool{}},
		{[]string{}, map[string]bool{}},
		{[]string{"a", "B"}, map[string]bool{"a": true, "B": true}},
		{[]string{"a", "a"}, map[string]bool{"a": true}},
	} {
		got := MakeSet(test.list)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf(
				"MakeSet(%v): got %v, want %v",
				test.list, got, test.want,
			)
		}
	}
}

func TestSortedList(t *testing.T) {
	for _, test := range []struct {
		set  map[string]bool
		want []string
	}{
		{nil, nil},
		{map[string]bool{}, nil},
		{map[string]bool{"a": true}, []string{"a"}},
		{map[string]bool{"b": true, "a": true, "c": true},
			[]string{"a", "b", "c"}},
		{map[string]bool{"B": true, "a": true}, []string{"B", "a"}},
	} {
		got := SortedList(test.set)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf(
				"SortedList(%v): got %v, want %v",
				test.set, got, test.want,
			)
		}
	}
}
