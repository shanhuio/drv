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

	const tmpName = "ca-certificates-202606.crt"
	tmpFile := path.Join("/tmp", tmpName)

	// Copy the new bundle into /tmp first, then overwrite the destination
	// in place with cp. The destination is a mounted file, so it cannot be
	// replaced with a move; it must be overwritten.
	s := tarutil.NewStream()
	s.AddBytes(tmpName, tarutil.ModeMeta(0644), caCertificates202606)
	if err := b.CopyInTarStream(s, "/tmp"); err != nil {
		return false, errcode.Annotate(err, "copy in new root CA certificates")
	}
	// Best-effort clean up of the staged file.
	defer func() {
		if _, err := b.ExecRet([]string{"rm", "-f", tmpFile}); err != nil {
			log.Printf("remove staged root CA certificates %q: %s", tmpFile, err)
		}
	}()

	ret, err := b.ExecRet([]string{"sudo", "cp", tmpFile, rootCACertFile})
	if err != nil {
		return false, errcode.Annotate(err, "overwrite root CA certificates")
	}
	if ret != 0 {
		return false, errcode.Internalf(
			"overwrite root CA certificates: exit %d", ret,
		)
	}

	log.Printf("replaced broken root CA certificates %q", rootCACertFile)
	return true, nil
}
