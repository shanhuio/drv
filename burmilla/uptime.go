package burmilla

import (
	"strconv"
	"strings"
	"time"

	"shanhu.io/g/errcode"
)

// Uptime returns the uptime of the system.
func Uptime(b *Burmilla) (time.Duration, error) {
	line, err := b.ExecOutput([]string{"cat", "/proc/uptime"})
	if err != nil {
		return 0, errcode.Annotate(err, "query system uptime")
	}

	fields := strings.Fields(string(line))
	if len(fields) != 2 {
		return 0, errcode.Internalf(
			"system uptime line has %d fields, want 2", len(fields),
		)
	}

	secs, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, errcode.Annotate(err, "parse uptime")
	}

	uptime := time.Duration(int64(secs*1e9)) * time.Nanosecond
	return uptime, nil
}
