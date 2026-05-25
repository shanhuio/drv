package jarvis

import (
	"crypto/sha256"
	"time"

	"shanhu.io/g/aries"
	"shanhu.io/g/errcode"
	"shanhu.io/g/oauth2"
	"shanhu.io/g/osutil"
	"shanhu.io/g/settings"
	"shanhu.io/g/signer"
	"shanhu.io/drv/drvapi"
	drvcfg "shanhu.io/drv/drvconfig"
	"shanhu.io/drv/homeapp"
)

type server struct {
	*backend
	drive       *drive
	appRegistry *appRegistry
	apps        *apps

	auth          *oauth2.Module
	sudoSessions  *sudoSessions
	loginSessions *loginSessions
	totp          *totp
	sshKeys       *sshKeys

	tmpls  *aries.Templates
	static *aries.StaticFiles

	updateSignal chan bool
}

func newServer(h *osutil.Home, c *drvcfg.Config) (*server, error) {
	back, err := newBackend(h)
	if err != nil {
		return nil, errcode.Annotate(err, "create backend")
	}

	rel := new(drvapi.Release)
	if err := back.settings.Get(keyBuild, rel); err != nil {
		if !errcode.IsNotFound(err) {
			return nil, errcode.Annotate(err, "read current build")
		}
	}
	appReg := newAppRegistry(rel)
	stateStore := &appsStateSettings{
		key:      keyAppsState,
		settings: back.settings,
	}
	if rel.Name != "" { // not first time install.
		if err := maybeSetAppsStateFromLegacy(stateStore, appReg); err != nil {
			return nil, errcode.Annotate(err, "set apps state from legacy")
		}
	}

	apps, err := newApps(&appsConfig{
		store:   stateStore,
		querier: appReg,
	})
	if err != nil {
		return nil, errcode.Annotate(err, "build apps control")
	}

	objs, err := newObjects(h.Var("objs"))
	if err != nil {
		return nil, errcode.Annotate(err, "create objects store")
	}

	kernel := &kernel{
		settings:    back.settings,
		appDomains:  back.appDomains,
		appRegistry: appReg,
		apps:        apps,
		objects:     objs,
	}
	drive, err := newDrive(c, kernel)
	if err != nil {
		return nil, err
	}

	if err := apps.setMaker(newBuiltInApps(drive)); err != nil {
		return nil, errcode.Annotate(err, "setup builtin app stubs")
	}

	sessionKey, err := settings.String(back.settings, keySessionHMAC)
	if err != nil {
		return nil, errcode.Annotate(err, "read session key")
	}
	sudoSessions := newSudoSessions(sessionKey)
	loginSessions := newLoginSessions(sessionKey)
	auth := oauth2.NewModule(&oauth2.Config{
		SessionKey: []byte(sessionKey),
		PreSignOut: func(c *aries.C) error {
			sudoSessions.ClearCookie(c)
			return nil
		},
	})

	signerKey := sha256.Sum256([]byte("state:" + sessionKey))
	stateSigner := signer.New(signerKey[:])

	totpCfg := &totpConfig{
		sudo:        sudoSessions,
		stateSigner: stateSigner,
		logs:        back.securityLogs,
		issuer: func() (string, error) {
			v, err := settings.String(back.settings, homeapp.KeyMainDomain)
			if errcode.IsNotFound(err) {
				return "unknown.homedrive.io", nil
			}
			return v, err
		},
		now: time.Now,
	}
	totp, err := newTOTP(back.users, totpCfg)
	if err != nil {
		return nil, errcode.Annotate(err, "create totp")
	}

	s := &server{
		backend:     back,
		drive:       drive,
		appRegistry: appReg,
		apps:        apps,

		auth:          auth,
		sudoSessions:  sudoSessions,
		loginSessions: loginSessions,
		totp:          totp,
		sshKeys:       newSSHKeys(drive),

		tmpls:  aries.NewTemplates(h.Lib("tmpl"), nil),
		static: aries.NewStaticFiles(h.Lib("static")),

		updateSignal: make(chan bool),
	}
	return s, nil
}

func (s *server) Drive() *drive { return s.drive }

func (s *server) f(f func(s *server, c *aries.C) error) aries.Func {
	return func(c *aries.C) error { return f(s, c) }
}
