package burmilla

import (
	"bytes"
	"io/ioutil"
	"strings"

	"shanhu.io/std/docker"
	"shanhu.io/std/tarutil"
)

// Burmilla provides the
type Burmilla struct {
	sysDock *docker.Client
}

// New creates a new burmilla stub.
func New(d *docker.Client) *Burmilla {
	return &Burmilla{sysDock: d}
}

// Console returns the console container.
func (b *Burmilla) Console() *docker.Cont {
	return docker.NewCont(b.sysDock, "console")
}

// ExecOutput executes a command on the OS's console
// and returns its output.
func (b *Burmilla) ExecOutput(args []string) ([]byte, error) {
	out := new(bytes.Buffer)
	c := b.Console()
	if err := execError(c.ExecWithSetup(&docker.ExecSetup{
		Cmd:    args,
		Stdout: out,
	})); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// ExecRet executes a command on the OS's console and returns its return
// value.
func (b *Burmilla) ExecRet(args []string) (int, error) {
	c := b.Console()
	return c.ExecWithSetup(&docker.ExecSetup{
		Cmd:    args,
		Stdout: ioutil.Discard,
	})
}

// CopyInTarStream copies files into the console's filesystem.
func (b *Burmilla) CopyInTarStream(s *tarutil.Stream, target string) error {
	c := b.Console()
	return docker.CopyInTarStream(c, s, target)
}

// ListOS lists the avaiable OS versions.
func ListOS(b *Burmilla) ([]string, error) {
	out, err := b.ExecOutput(strings.Fields("ros os list"))
	if err != nil {
		return nil, err
	}
	s := string(out)
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}
