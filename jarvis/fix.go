package jarvis

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"log"
	"path"

	"shanhu.io/drv/burmilla"
	"shanhu.io/drv/homeapp/nextcloud"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
	"shanhu.io/std/tarutil"
)

// caCertificates202606 is the replacement root CA certificate bundle that
// overwrites the known-broken one on the OS console.
//
//go:embed ca-certificates-202606.crt
var caCertificates202606 []byte

const (
	// rootCACertFile is the root CA bundle Rancher/Burmilla ships on the
	// console container.
	rootCACertFile = "/etc/ssl/certs/ca-certificates.crt.rancher"

	// brokenRootCACertSHA256 is the sha256 sum of the broken bundle that
	// needs to be replaced.
	brokenRootCACertSHA256 = "7c913c3f91405559e2a5b9b93e2eb20a112bea02e020797f58911caf2a6794ea"
)

// fixRootCACertificates replaces the OS console's root CA certificate bundle
// when it matches the known-broken version.
func fixRootCACertificates(b *burmilla.Burmilla) error {
	cur, err := docker.ReadContFile(b.Console(), rootCACertFile)
	if err != nil {
		return errcode.Annotate(err, "read root CA certificates")
	}

	sum := sha256.Sum256(cur)
	if hex.EncodeToString(sum[:]) != brokenRootCACertSHA256 {
		return nil // not the known-broken bundle; leave it alone
	}

	s := tarutil.NewStream()
	s.AddBytes(path.Base(rootCACertFile), tarutil.ModeMeta(0644), caCertificates202606)
	if err := b.CopyInTarStream(s, path.Dir(rootCACertFile)); err != nil {
		return errcode.Annotate(err, "overwrite root CA certificates")
	}

	log.Printf("replaced broken root CA certificates %q", rootCACertFile)
	return nil
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
