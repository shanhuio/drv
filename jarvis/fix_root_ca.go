package jarvis

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"log"
	"path"

	"shanhu.io/drv/burmilla"
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
// when it matches the known-broken version. It returns true when the bundle
// was overwritten.
func fixRootCACertificates(b *burmilla.Burmilla) (bool, error) {
	cur, err := docker.ReadContFile(b.Console(), rootCACertFile)
	if err != nil {
		return false, errcode.Annotate(err, "read root CA certificates")
	}

	sum := sha256.Sum256(cur)
	if hex.EncodeToString(sum[:]) != brokenRootCACertSHA256 {
		return false, nil // not the known-broken bundle; leave it alone
	}

	s := tarutil.NewStream()
	s.AddBytes(path.Base(rootCACertFile), tarutil.ModeMeta(0644), caCertificates202606)
	if err := b.CopyInTarStream(s, path.Dir(rootCACertFile)); err != nil {
		return false, errcode.Annotate(err, "overwrite root CA certificates")
	}

	log.Printf("replaced broken root CA certificates %q", rootCACertFile)
	return true, nil
}
