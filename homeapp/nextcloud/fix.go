package nextcloud

import (
	"io"

	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
	"shanhu.io/g/settings"
	"shanhu.io/drv/semver"
)

func fix(cont *dock.Cont, s settings.Settings) error {
	version, err := readTrueVersion(cont)
	if err != nil {
		return errcode.Annotate(err, "get version")
	}
	major, err := semver.Major(version)
	if err != nil {
		return errcode.Add(errcode.Internal, err)
	}
	return fixVersion(cont, s, major)
}

func fixVersion(cont *dock.Cont, s settings.Settings, major int) error {
	if major >= 30 {
		// For version 30+, this needs to be executed every time a new
		// docker is installed.
		if err := aptUpdate(cont, io.Discard); err != nil {
			return errcode.Annotate(err, "apt update for nextcloud30+")
		}

		pkgs := []string{
			"smbclient",
			"libsmbclient-dev",
		}
		if err := aptInstall(cont, pkgs, io.Discard); err != nil {
			return errcode.Annotate(err, "install additional packages")
		}

		if err := enableSMB(cont, io.Discard); err != nil {
			return errcode.Annotate(err, "enable SMB")
		}
	}

	if err := setCronMode(cont); err != nil {
		return errcode.Annotate(err, "set cron mode")
	}

	// The following fixes might be also needed in minor upgrades.
	for _, cmd := range []string{
		"db:add-missing-indices",
		"db:convert-filecache-bigint",
	} {
		if _, err := occOutput(cont, []string{cmd, "-n"}); err != nil {
			return errcode.Annotate(err, cmd)
		}
	}

	k := fixKey(major)
	if k == "" {
		return nil
	}
	ok, err := s.Has(k)
	if err != nil {
		return errcode.Annotatef(err, "check fixed flag v%d", major)
	}
	if ok {
		return nil
	}

	for _, cmd := range []string{
		"db:add-missing-columns",
		"db:add-missing-primary-keys",
	} {
		if _, err := occOutput(cont, []string{cmd, "-n"}); err != nil {
			return errcode.Annotate(err, cmd)
		}
	}

	/* Maybe include this next time?

	if major >= 30 {
		// Also perform heavy migrations.
		cmd := []string{"maintenance:repair", "--include-expensive"}
		if _, err := occOutput(cont, cmd); err != nil {
			return errcode.Annotate(err, strings.Join(cmd, " "))
		}
	}

	*/

	if err := s.Set(k, true); err != nil {
		return errcode.Annotatef(err, "set fixed flag v%d", major)
	}
	return nil
}
