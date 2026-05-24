package burmilla

import (
	"bytes"
	"io/ioutil"
	"strings"

	"shanhu.io/g/dock"
	"shanhu.io/g/tarutil"
)

// Burmilla provides the
type Burmilla struct {
	sysDock *dock.Client
}

// New creates a new burmilla stub.
func New(d *dock.Client) *Burmilla {
	return &Burmilla{sysDock: d}
}

// Console returns the console container.
func (b *Burmilla) Console() *dock.Cont {
	return dock.NewCont(b.sysDock, "console")
}

// ExecOutput executes a command on the OS's console
// and returns its output.
func (b *Burmilla) ExecOutput(args []string) ([]byte, error) {
	out := new(bytes.Buffer)
	c := b.Console()
	if err := execError(c.ExecWithSetup(&dock.ExecSetup{
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
	return c.ExecWithSetup(&dock.ExecSetup{
		Cmd:    args,
		Stdout: ioutil.Discard,
	})
}

// CopyInTarStream copies files into the console's filesystem.
func (b *Burmilla) CopyInTarStream(s *tarutil.Stream, target string) error {
	c := b.Console()
	return dock.CopyInTarStream(c, s, target)
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
