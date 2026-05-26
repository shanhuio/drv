package nextcloud

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"shanhu.io/drv/executil"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
)

func configSaysInstalled(config []byte) bool {
	// check if the config file has the `'installed' => true` line.
	lines := strings.Split(string(config), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == `'installed' => true,` {
			return true
		}
		if line == `'installed' => false,` {
			return false
		}
	}
	return false
}

func exec(c *docker.Cont, cmd []string, out io.Writer) error {
	return executil.RetError(c.ExecWithSetup(&docker.ExecSetup{
		Cmd:    cmd,
		Stdout: out,
	}))
}

func aptUpdate(c *docker.Cont, out io.Writer) error {
	cmd := []string{"apt-get", "update"}
	return exec(c, cmd, out)
}

func aptInstall(c *docker.Cont, pkgs []string, out io.Writer) error {
	cmd := []string{"apt-get", "install", "-y"}
	cmd = append(cmd, pkgs...)
	return exec(c, cmd, out)
}

func enableSMB(c *docker.Cont, out io.Writer) error {
	peclList := new(bytes.Buffer)
	if err := exec(c, []string{"pecl", "list"}, peclList); err != nil {
		return errcode.Annotate(err, "pecl list")
	}
	lines := strings.Split(peclList.String(), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) == 3 && fields[0] == "smbclient" {
			// smbclient already installed; let's skip.
			return nil
		}
	}

	cmd := []string{"pecl", "install", "smbclient"}
	if err := exec(c, cmd, out); err != nil {
		return errcode.Annotate(err, "pecl install")
	}
	cmd = []string{"docker-php-ext-enable", "smbclient"}
	if err := exec(c, cmd, out); err != nil {
		return errcode.Annotate(err, "docker-php-ext-enable")
	}
	return nil
}

func occRet(
	c *docker.Cont, args []string, out io.Writer,
) (int, error) {
	cmd := append([]string{"php", "occ"}, args...)
	return c.ExecWithSetup(&docker.ExecSetup{
		Cmd:    cmd,
		User:   "www-data",
		Stdout: out,
	})
}

func occ(c *docker.Cont, args []string, out io.Writer) error {
	return executil.RetError(occRet(c, args, out))
}

func occOutput(c *docker.Cont, args []string) ([]byte, error) {
	out := new(bytes.Buffer)
	if err := occ(c, args, out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func testReadConfig(cont *docker.Cont) ([]byte, error) {
	const configFile = "/var/www/html/config/config.php"

	ret, err := cont.ExecWithSetup(&docker.ExecSetup{
		Cmd:  []string{"/usr/bin/test", "-e", configFile},
		User: "www-data",
	})
	if err != nil {
		return nil, errcode.Annotate(err, "test config.php")
	}
	if ret != 0 {
		return nil, nil
	}
	return docker.ReadContFile(cont, configFile)
}

func cron(cont *docker.Cont) error {
	return executil.RetError(cont.ExecWithSetup(&docker.ExecSetup{
		Cmd:    []string{"php", "cron.php"},
		User:   "www-data",
		Stdout: io.Discard,
	}))
}

func fixKey(major int) string {
	if major >= 18 && major < 10000 {
		return fmt.Sprintf("nextcloud-%d-fixed", major)
	}
	return ""
}

func setRedisPassword(cont *docker.Cont, pwd string) error {
	// TODO(h8liu): should first check if redis password is incorrect.
	args := []string{
		"config:system:set", "-q",
		"--value=" + pwd,    // value
		"redis", "password", // key
	}
	return occ(cont, args, nil)
}

func setCronMode(cont *docker.Cont) error {
	args := []string{
		"config:app:set", "-q", "--value=cron",
		"core", "backgroundjobs_mode",
	}
	return occ(cont, args, nil)
}
