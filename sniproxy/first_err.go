package sniproxy

type firstErr struct {
	err error
}

func (e *firstErr) set(err error) {
	if err == nil {
		return
	}
	if e.err == nil {
		e.err = err
	}
}

func (e *firstErr) get() error {
	return e.err
}
