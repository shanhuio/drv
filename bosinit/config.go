package bosinit

import (
	"bytes"
	"io"

	yaml "gopkg.in/yaml.v2"
)

// Config is the root of the cloud-init configuration YAML.
type Config struct {
	Rancher    *Rancher     `yaml:",omitempty"`
	WriteFiles []*WriteFile `yaml:"write_files,omitempty"`

	SSHAuthorizedKeys []string `yaml:"ssh_authorized_keys,omitempty"`
}

func (c *Config) encodeInto(w io.Writer) error {
	enc := yaml.NewEncoder(w)
	if err := enc.Encode(c); err != nil {
		return err
	}
	return enc.Close()
}

// CloudConfig encodes the cloud-init config into YAML form.
// It also has "#cloud-config" shebang on the first line.
func (c *Config) CloudConfig() ([]byte, error) {
	buf := new(bytes.Buffer)
	io.WriteString(buf, "#cloud-config\n")
	if err := c.encodeInto(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Encode encodes the cloud-init config into YAML form.
func (c *Config) Encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := c.encodeInto(buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ParseConfig parses an Rancher/Burmilla OS init config.
func ParseConfig(bs []byte) (*Config, error) {
	c := new(Config)
	if err := yaml.Unmarshal(bs, c); err != nil {
		return nil, err
	}
	return c, nil
}
