package homerelease

import (
	"encoding/json"

	"shanhu.io/drv/drvapi"
	"shanhu.io/g/creds"
	"shanhu.io/g/errcode"
	"shanhu.io/g/jsonutil"
)

func cmdPush(server string, args []string) error {
	flags := cmdFlags.New()
	objs := flags.String(
		"objs", "out/docker/homedrv/objs.tar", "path to objects tarball",
	)
	rel := flags.String(
		"release", "out/docker/homedrv/release.json", "path to release info",
	)
	user := flags.String(
		"user", "root", "user to call the push API",
	)
	_ = flags.ParseArgs(args)

	c, err := creds.DialAsUser(*user, server)
	if err != nil {
		return errcode.Annotate(err, "dial server")
	}

	up := &Uploader{
		Client:  c,
		DataURL: "/obj",
		APIURL:  "/api/obj",
	}
	if err := up.Upload(*objs); err != nil {
		return errcode.Annotate(err, "upload objects")
	}

	release := new(drvapi.Release)
	if err := jsonutil.ReadFile(*rel, release); err != nil {
		return errcode.Annotate(err, "read release file")
	}
	newName, err := MakeReleaseName(release.Type)
	if err != nil {
		return errcode.Annotate(err, "make release name")
	}
	release.Name = newName

	bs, err := json.Marshal(release)
	if err != nil {
		return errcode.Annotate(err, "marshal release")
	}
	return c.Call("/api/admin/push-update", bs, nil)
}
