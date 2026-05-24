package doorway

import (
	"context"
	"os"
)

// Identity provides an identity for dialing fabrics.
type Identity interface {
	// Load loads the identity private key. Returns errcode.NotFound error
	// if key is not yet provisioned.
	Load(ctx context.Context) ([]byte, error)
}

type staticIdentity struct {
	pri []byte
}

func newStaticIdentity(bs []byte) *staticIdentity {
	return &staticIdentity{bs}
}

func newFileIdentity(f string) (*staticIdentity, error) {
	bs, err := os.ReadFile(f)
	if err != nil {
		return nil, err
	}
	return newStaticIdentity(bs), nil
}

func (s *staticIdentity) Load(ctx context.Context) ([]byte, error) {
	cp := make([]byte, len(s.pri))
	copy(cp, s.pri)
	return cp, nil
}

// NewFileIdentity loads a private key from a file.
func NewFileIdentity(f string) (Identity, error) {
	id, err := newFileIdentity(f)
	if err != nil {
		return nil, err
	}
	return id, nil
}
