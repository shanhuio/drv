package jarvis

import (
	"time"

	"shanhu.io/g/errcode"
)

const failureWindow = 3 * time.Minute

type recentFailures struct {
	Timestamps []int64 `json:",omitempty"`
}

func (r *recentFailures) update(now time.Time) {
	cut := now.Add(-failureWindow).UnixNano()
	ts := make([]int64, 0, 5)
	for _, t := range r.Timestamps {
		if t > cut {
			ts = append(ts, t)
		}
	}
	r.Timestamps = ts
}

func (r *recentFailures) count(now time.Time) int {
	r.update(now)
	return len(r.Timestamps)
}

func (r *recentFailures) add(now time.Time) int {
	r.update(now)
	r.Timestamps = append(r.Timestamps, now.UnixNano())
	return len(r.Timestamps)
}

func (r *recentFailures) clear() {
	r.Timestamps = nil
}

var errTooManyFailures = errcode.Unauthorizedf("too many recent failures")
