package burmilla

import (
	"strconv"
	"strings"

	"shanhu.io/g/errcode"
)

func parseIDOutput(bs []byte) (int, error) {
	return strconv.Atoi(strings.TrimSpace(string(bs)))
}

// UserID returns the uid of a particular user.
func UserID(b *Burmilla, name string) (int, error) {
	out, err := b.ExecOutput([]string{"id", "-u", name})
	if err != nil {
		return 0, errcode.Annotate(err, "get user id")
	}
	id, err := parseIDOutput(out)
	if err != nil {
		return 0, errcode.Annotate(err, "parse user id")
	}
	return id, nil
}

// GroupID returns the gid of a particular user
func GroupID(b *Burmilla, name string) (int, error) {
	out, err := b.ExecOutput([]string{"id", "-g", name})
	if err != nil {
		return 0, errcode.Annotate(err, "get group id")
	}
	id, err := parseIDOutput(out)
	if err != nil {
		return 0, errcode.Annotate(err, "parse group id")
	}
	return id, nil
}
