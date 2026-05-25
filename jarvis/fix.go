package jarvis

import (
	"log"

	"shanhu.io/g/errcode"
	"shanhu.io/drv/homeapp/nextcloud"
)

func fixThings(d *drive) {
	if err := fixOSUpgradeURL(d); err != nil {
		log.Println("fix os upgrade url: ", err)
	}
	if d.apps.isInstalled(nextcloud.Name) {
		if err := nextcloud.Fix(d); err != nil {
			log.Println("fix nextcloud:", err)
		}
	}
}

func fixOSUpgradeURL(d *drive) error {
	if !isOSUpdateSupported(d) {
		return nil
	}
	b, err := d.burmilla()
	if err != nil {
		return errcode.Annotate(err, "init os stub")
	}
	return setOSUpdateSource(b)
}
