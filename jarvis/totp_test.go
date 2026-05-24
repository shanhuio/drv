package jarvis

import (
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	totppkg "github.com/pquerna/otp/totp"
	"shanhu.io/g/aries"
	"shanhu.io/g/httputil"
	"shanhu.io/g/pisces"
	"shanhu.io/g/signer"
)

type noopChecker struct{}

func (m *noopChecker) Check(_ *aries.C) error {
	// Always pass.
	return nil
}

func TestTOTPFlow(t *testing.T) {
	const testUser = "test-user"
	const testPassword = "123456"

	tables := pisces.NewTables(nil) // In-memory table.
	users := newUsers(tables)
	users.create(testUser, testPassword)

	signer := signer.New([]byte("test-key"))

	// Hardcode a timestamp that is not close to minute boundary.
	now := time.Date(1972, time.October, 25, 19, 21, 28, 0, time.UTC)

	totp, err := newTOTP(users, &totpConfig{
		sudo:        &noopChecker{},
		stateSigner: signer,
		logs:        nil,
		issuer:      nil,
		now: func() time.Time {
			return now
		},
	})
	if err != nil {
		t.Fatal("create totp", err)
	}

	r := aries.NewRouter()
	r.DirService("totp", totp.api())

	s := httptest.NewServer(aries.Func(func(c *aries.C) error {
		c.User = testUser
		return r.Serve(c)
	}))
	defer s.Close()

	c, err := httputil.NewClient(s.URL)
	if err != nil {
		t.Fatal("new client", err)
	}

	// Check enable.
	setupReq := &SetupTOTPRequest{}
	totpSetup := new(TOTPSetup)
	if err := c.Call("totp/setup", setupReq, totpSetup); err != nil {
		t.Fatal("fail to enable TOTP", err)
	}

	// Read secret out of key URL.
	totpURL, err := url.Parse(totpSetup.URL)
	if err != nil {
		t.Fatal("fail to parse TOTP key URL", err)
	}
	secret := totpURL.Query().Get("secret")

	// Generate a passcode so we can activate.
	otp, err := totppkg.GenerateCodeCustom(
		secret,
		now,
		totppkg.ValidateOpts{
			Digits:    totpDigits,
			Algorithm: totpAlgorithm,
		},
	)
	if err != nil {
		t.Fatal("can not generate otp", err)
	}

	// Advance clock by 1 second.
	now = now.Add(time.Second)
	enableReq := &EnableTOTPRequest{
		SignedSecret: totpSetup.SignedSecret,
		OTP:          otp,
	}
	// Successfully activate.
	if err := c.Call("totp/enable", enableReq, nil); err != nil {
		t.Error("fail to enable totp", err)
	}

	// Successfully disable.
	disableReq := &DisableTOTPRequest{}
	if err := c.Call("totp/disable", disableReq, nil); err != nil {
		t.Error("fail to disable totp", err)
	}
}
