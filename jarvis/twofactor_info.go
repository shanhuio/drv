package jarvis

type totpInfo struct {
	Secret string
}

type twoFactorInfo struct {
	TOTP *totpInfo `json:",omitempty"`
}
