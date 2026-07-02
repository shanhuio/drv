package jarvis

import (
	"flag"
	"log"
	"time"

	drvcfg "shanhu.io/drv/drvconfig"
	"shanhu.io/drv/homeboot"
	"shanhu.io/g/aries"
	"shanhu.io/g/osutil"
	"shanhu.io/std/errcode"

	_ "github.com/lib/pq"  // for postgres
	_ "modernc.org/sqlite" // sqlite db driver
)

// Main is the main entrance of jarvis server or client program.
func Main() {
	if osutil.Arg0Base() == "jarvisd" {
		serverMain()
		return
	}
	clientMain()
}

func makeService(s *server, api aries.Service) aries.Service {
	return &aries.ServiceSet{
		Auth:  s.auth.Auth(),
		User:  userRouter(s, api),
		Guest: guestRouter(s),
	}
}

func bg(s *server) {
	d := s.Drive()

	// Before starting the system tasks scheduler, make sure the system is
	// properlly installed.
	installed, err := d.settings.Has(keyBuild)
	if err != nil {
		// Basic install check failed.
		log.Println("check installed:", err)
	} else if !installed { // This is first time.
		if err := downloadAndInstall(d); err != nil {
			log.Println("install failed:", err)
		}
	} else { // Not first time.
		if err := maybeFinishUpdate(d); err != nil {
			log.Println("update failed:", err)
			// It is important to proceed here, as the next update might be
			// able to fix the issue. At this point, the apps are in
			// undefiend state, but jarvis is already on the latest.
		}
		fixThings(d)
	}

	if d.config.Channel != "" {
		// Subscribe channel and maybe schedule update task.
		go cronUpdateOnChannel(d, s.updateSignal)
	}

	go cronNextcloud(d)

	d.tasks.bg() // Handle background system tasks now.
}

func run(homeDir, addr string) error {
	h, err := osutil.NewHome(homeDir)
	if err != nil {
		return errcode.Annotate(err, "open home dir")
	}

	// jarvis reads config from var.
	config, err := readConfig(h)
	if err != nil {
		return errcode.Annotate(err, "read config")
	}

	s, err := newServer(h, config)
	if err != nil {
		return errcode.Annotate(err, "create server")
	}
	empty := drvcfg.Image(config.Naming, "empty")
	if err := homeboot.BuildEmptyIfNotExist(s.drive.dock, empty); err != nil {
		return errcode.Annotate(err, "build empty docker image")
	}

	if !config.External {
		if err := killOldCoreIfExist(s.drive); err != nil {
			return errcode.Annotate(err, "kill old core")
		}
	}
	d := s.Drive()
	if err := maybeUpdateOS(d); err != nil {
		// This step might fail if the OS is in a weird process
		// We do not return an error here; otherwise fabrics will be trapped
		// in a crash-loop that cannot pick up new updates.
		log.Printf("ERROR: update os failed: %s", err)
		time.Sleep(10 * time.Second)
	}

	go bg(s)

	const sock = "var/jarvis.sock"
	log.Printf("serve on %s and %s", sock, addr)

	api := apiRouter(s)
	go func(api aries.Service) {
		r := aries.NewRouter()
		r.DirService("api", api)

		if err := aries.ListenAndServe(sock, r); err != nil {
			log.Fatal(errcode.Annotate(err, "listen and serve on socket"))
		}
	}(api)

	service := makeService(s, api)
	if err := aries.ListenAndServe(addr, service); err != nil {
		return errcode.Annotate(err, "listen and serve")
	}
	return nil
}

func serverMain() {
	addr := flag.String("addr", "localhost:3377", "address to listen on")
	home := flag.String("home", ".", "home dir")
	flag.Parse()

	if err := run(*home, *addr); err != nil {
		log.Fatal(err)
	}
}
