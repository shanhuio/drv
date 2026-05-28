package sniproxy

import (
	"net/url"

	"encoding/json"

	"shanhu.io/g/errcode"
)

// Options is a the JSON marshalable options for dialing an endpoint.
type Options struct {
	// Using a new websocket connection for each new incoming
	// new connection.
	Siding bool `json:",omitempty"`

	// Remote enables sending remote address.
	DialWithAddr bool `json:",omitempty"`
}

func decodeOptions(s string) (*Options, error) {
	opt := new(Options)
	if s == "" {
		return opt, nil
	}
	if err := json.Unmarshal([]byte(s), opt); err != nil {
		return nil, errcode.InvalidArgf("invalid options: %s", err)
	}

	return opt, nil
}

func optionsFromQuery(q url.Values) (*Options, error) {
	return decodeOptions(q.Get("opt"))
}
