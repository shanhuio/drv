package jarvis

import (
	"encoding/json"
	"time"

	"shanhu.io/g/rand"
)

// LogEntry is a log entry for jarvis to display.
type LogEntry struct {
	K    string
	T    int64
	User string `json:",omitempty"`
	Text string `json:",omitempty"`
	Type string `json:",omitempty"`
	V    []byte `json:",omitempty"`

	// The following fields are only used in Javascript.
	TSec int64  `json:",omitempty"`
	VStr string `json:",omitempty"`
}

func newLogEntryAt(t time.Time, user, text string) *LogEntry {
	t = t.UTC()
	k := t.Format(time.RFC3339Nano) + "-" + rand.Letters(6)
	return &LogEntry{
		K:    k,
		T:    t.UnixNano(),
		User: user,
		Text: text,
	}
}

func newLogEntry(user, text string) *LogEntry {
	return newLogEntryAt(time.Now(), user, text)
}

func (e *LogEntry) setJSONValue(typ string, v interface{}) error {
	bs, err := json.Marshal(v)
	if err != nil {
		return err
	}
	e.Type = typ
	e.V = bs
	return nil
}

const (
	logTypeLoginAttempt   = "loginAttempt"
	logTypeTwoFactorEvent = "twoFactorEvent"
	logTypeChangePassword = "changePassword"
)
