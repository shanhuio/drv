package jarvis

import (
	"fmt"

	"shanhu.io/g/pisces"
	"shanhu.io/std/errcode"
)

type securityLogs struct {
	t *pisces.KV
}

func newSecurityLogs(b *pisces.Tables) *securityLogs {
	return &securityLogs{t: b.NewOrderedKV("security_logs")}
}

func (b *securityLogs) add(entry *LogEntry) error {
	return b.t.Add(entry.K, entry)
}

type loginEvent struct {
	From      string `json:",omitempty"`
	TwoFactor string `json:",omitempty"`
	Failed    bool   `json:",omitempty"`
}

func (b *securityLogs) recordLogin(user, from, twoFactor string) error {
	msg := fmt.Sprintf("login from %q", from)
	entry := newLogEntry(user, msg)
	if err := entry.setJSONValue(logTypeLoginAttempt, &loginEvent{
		From:      from,
		TwoFactor: twoFactor,
	}); err != nil {
		return errcode.Annotate(err, "set log value")
	}
	return b.add(entry)
}

func (b *securityLogs) recordFailedLogin(
	user, from, twoFactor string,
) error {
	msg := fmt.Sprintf("failed login from %q", from)
	entry := newLogEntry(user, msg)
	if err := entry.setJSONValue(logTypeLoginAttempt, &loginEvent{
		From:      from,
		TwoFactor: twoFactor,
		Failed:    true,
	}); err != nil {
		return errcode.Annotate(err, "set log value")
	}
	return b.add(entry)
}

type changePasswordEvent struct{}

func (b *securityLogs) recordChangePassword(user string) error {
	msg := fmt.Sprintf("password of user %q changed", user)
	entry := newLogEntry(user, msg)
	ev := &changePasswordEvent{}
	if err := entry.setJSONValue(logTypeChangePassword, ev); err != nil {
		return errcode.Annotate(err, "set log value")
	}
	return b.add(entry)
}

const methodTOTP = "TOTP"

type twoFactorEvent struct {
	Method string `json:",omitempty"`
	Event  string `json:",omitempty"`
}

func (ev *twoFactorEvent) String() string {
	return fmt.Sprintf("two factor auth: %s - %s", ev.Method, ev.Event)
}

func (b *securityLogs) recordTwoFactorEvent(user, m, event string) error {
	ev := &twoFactorEvent{
		Method: m,
		Event:  event,
	}
	entry := newLogEntry(user, ev.String())
	if err := entry.setJSONValue(logTypeTwoFactorEvent, ev); err != nil {
		return errcode.Annotate(err, "set log value")
	}
	return b.add(entry)
}

func (b *securityLogs) list(page int) ([]*LogEntry, error) {
	if page < 0 {
		page = 0
	}
	const perPage = 100
	partial := &pisces.KVPartial{
		Offset: uint64(page) * perPage,
		N:      perPage,
		Desc:   true,
	}
	var entries []*LogEntry
	it := &pisces.Iter{
		Make: func() any { return new(LogEntry) },
		Do: func(_ string, v any) error {
			entries = append(entries, v.(*LogEntry))
			return nil
		},
	}
	if err := b.t.WalkPartial(partial, it); err != nil {
		return nil, err
	}
	return entries, nil
}
