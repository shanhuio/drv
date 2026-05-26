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
