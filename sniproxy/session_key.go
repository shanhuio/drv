package sniproxy

import (
	"encoding/json"
)

type sessionKey struct {
	ID  uint64
	Key uint64
}

func (k *sessionKey) encode() (string, error) {
	bs, err := json.Marshal(k)
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

func decodeSessionKey(s string) (*sessionKey, error) {
	k := new(sessionKey)
	if err := json.Unmarshal([]byte(s), k); err != nil {
		return nil, err
	}
	return k, nil
}
