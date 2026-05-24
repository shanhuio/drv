package drvapi

// UserSSHKey saves a user's SSH public key.
type UserSSHKey struct {
	ID        string
	PublicKey string

	TimeCreatedSec int64 // Timestamp in seconds.
	TimeCreated    int64 // Timestamp in seconds, legacy.
}

// UserSSHKeys wraps a list of user SSH keys.
type UserSSHKeys struct {
	Keys []*UserSSHKey `json:",omitempty"`
}

// UserSSHKeyLines is a list of user SSH keys.
type UserSSHKeyLines struct {
	Keys []string
}
