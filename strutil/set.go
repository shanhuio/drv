package strutil

import (
	"sort"
)

// MakeSet converts a list of strings to a set of strings.
func MakeSet(lst []string) map[string]bool {
	ret := make(map[string]bool)
	for _, s := range lst {
		ret[s] = true
	}
	return ret
}

// SortedList returns the sorted list of a set of strings.
func SortedList(set map[string]bool) []string {
	var ret []string
	for s := range set {
		ret = append(ret, s)
	}
	sort.Strings(ret)
	return ret
}
