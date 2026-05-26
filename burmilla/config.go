package burmilla

import (
	"bytes"
	"fmt"
	"strings"

	"shanhu.io/g/bosinit"
	"shanhu.io/g/strutil"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
)

// ConfigExport exports the Bumilla OS's configuration.
func ConfigExport(b *Burmilla) (*bosinit.Config, error) {
	out, err := b.ExecOutput(strings.Fields("ros config export"))
	if err != nil {
		return nil, err
	}
	return bosinit.ParseConfig([]byte(out))
}

// ConfigMergeBad merges the config into the OS's configuration.
// Not working yet.
func ConfigMergeBad(b *Burmilla, config *bosinit.Config) error {
	bs, err := config.CloudConfig()
	if err != nil {
		return errcode.Annotate(err, "encode new cloud config")
	}
	in := bytes.NewReader(bs)

	c := b.Console()
	if err := execError(c.ExecWithSetup(&docker.ExecSetup{
		Cmd:   strings.Fields("ros config merge"),
		Stdin: in,
	})); err != nil {
		return errcode.Annotate(err, "ros config merge")
	}
	return nil
}

// ConfigSet modifies the OS's configuration.
func ConfigSet(b *Burmilla, k, v string) error {
	if _, err := b.ExecOutput([]string{
		"ros", "config", "set", k, v,
	}); err != nil {
		return errcode.Annotatef(err, "ros config set %q=%q", k, v)
	}
	return nil
}

// ConfigGet a particular field from the OS's configuration.
func ConfigGet(b *Burmilla, k string) (string, error) {
	bs, err := b.ExecOutput([]string{"ros", "config", "get", k})
	if err != nil {
		return "", errcode.Annotatef(err, "get config %q", k)
	}
	return strings.TrimSuffix(string(bs), "\n"), nil
}

// AddSSHKeys adds SSH key into the OS's configuration.
func AddSSHKeys(b *Burmilla, keys []string) error {
	config, err := ConfigExport(b)
	if err != nil {
		return err
	}

	keySet := strutil.MakeSet(config.SSHAuthorizedKeys)
	for _, k := range keys {
		if keySet[k] { // already have this
			continue
		}
		config.SSHAuthorizedKeys = append(
			config.SSHAuthorizedKeys, k,
		)
	}

	newConfig := &bosinit.Config{
		SSHAuthorizedKeys: config.SSHAuthorizedKeys,
	}
	return ConfigMerge(b, newConfig)
}

func quoteBashString(s string) string {
	// Borrowed from github.com/alessio/shellescape.
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

// ConfigMerge merges a configuration into the existing one.
func ConfigMerge(b *Burmilla, config *bosinit.Config) error {
	// TODO(h8liu): stream config via stdin and deprecate this.
	bs, err := config.CloudConfig()
	if err != nil {
		return errcode.Annotate(err, "encode cloud config")
	}

	script := fmt.Sprintf(
		"echo %s | sudo ros config merge", quoteBashString(string(bs)))
	return execError(b.ExecRet([]string{
		"/bin/bash", "-c", script,
	}))
}

func isOnDigitalOcean(config *bosinit.Config) bool {
	r := config.Rancher
	if r == nil {
		return false
	}
	init := r.CloudInit
	if init == nil {
		return false
	}

	for _, src := range init.DataSources {
		if src == "digitalocean" {
			return true
		}
	}
	return false
}
