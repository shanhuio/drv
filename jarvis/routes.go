package jarvis

import (
	"shanhu.io/g/aries"
	"shanhu.io/g/identity"
)

func guestRouter(s *server) *aries.Router {
	r := aries.NewRouter()

	r.Index(func(c *aries.C) error { return serveIndex(s, c) })

	dash := s.f(serveDashboard)
	r.Get("overview", dash)
	r.Get("ssh-keys", dash)
	r.Get("security-logs", dash)
	r.Get("change-password", dash)
	r.Get("2fa", dash)
	r.Get("2fa/enable-totp", dash)
	r.Get("2fa/disable-totp", dash)

	r.File("login", s.f(serveLogin))
	r.File("confirm-password", s.f(serveConfirmPassword))
	r.File("sudo", s.f(serveSudo))
	r.File("input-totp", s.f(serveInputTOTP))
	r.File("totp", s.f(serveCheckTOTP))

	static := s.static.Serve
	r.Get("style.css", static)
	r.Get("favicon.ico", static)
	r.Dir("js", static)
	r.Dir("jslib", static)
	r.Dir("img", static)
	r.Dir("fonts", static)

	return r
}

func userRouter(s *server, api aries.Service) *aries.Router {
	r := aries.NewRouter()
	r.DirService("api", api)
	r.DirService("obj", s.drive.objects)
	return r
}

func apiRouter(s *server) *aries.Router {
	r := aries.NewRouter()
	r.DirService("user", s.users.api())
	r.DirService("totp", s.totp.api())
	r.DirService("sshkeys", s.sshKeys.api())
	r.DirService("dashboard", dashboardAPI(s))
	r.DirService("id", identity.NewService(s.identity))
	r.DirService("obj", s.drive.objects.api())

	// All users are admin for now.
	r.DirService("admin", adminTasksAPI(s))

	return r
}
