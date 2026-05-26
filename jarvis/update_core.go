package jarvis

import (
	"log"
	"time"

	"shanhu.io/drv/homeapp/apputil"
	"shanhu.io/drv/homeboot"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
)

func killOldCoreIfExist(d *drive) error {
	cont := docker.NewCont(d.dock, d.oldCore())
	ok, err := cont.Exists()
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	if err := cont.Drop(); err != nil {
		if rmError := cont.ForceRemove(); rmError != nil {
			log.Println("force remove old core: ", rmError)
		}
		return err
	}
	return nil
}

func restartAs(d *drive, img string) error {
	// This is normally not necessary; just to make sure rename will succeed.
	if err := killOldCoreIfExist(d); err != nil {
		return errcode.Annotate(err, "kill old core")
	}
	if err := docker.RenameCont(
		d.dock, d.core(), d.oldCore(),
	); err != nil {
		return errcode.Annotate(err, "rename core to old")
	}

	hasSysDock := true
	if err := homeboot.CheckSystemDock(); err != nil {
		if !errcode.IsNotFound(err) {
			return errcode.Annotate(err, "check system docker")
		}
		hasSysDock = false
	}

	config := &homeboot.CoreConfig{
		Drive:       d.config,
		Image:       img,
		BindSysDock: hasSysDock,
	}
	id, err := homeboot.StartCore(d.dock, config)
	if err != nil {
		return err
	}
	log.Printf("new core started as %q", id)

	for { // Waiting to be killed.
		time.Sleep(time.Hour)
	}
	// unreachable.
}

func updateCore(d *drive, img string) error {
	// This is used in self-update in background, so this must be using
	// volumes already.
	self, err := docker.InspectCont(d.dock, d.core())
	if err != nil {
		return errcode.Annotate(err, "inspect self")
	}
	if self.Image == img {
		// Already up-to-date, no need to do anything.
		return apputil.ErrSameImage
	}
	return restartAs(d, img)
}
