package burmilla

import (
	"bytes"
	"text/template"

	"shanhu.io/std/errcode"
)

const mkdirCmdTmpl = `
[[ -d {{.Dir}} ]] || 
(mkdir -m 0700 {{.Dir}} && chown {{.User}}:{{.User}} {{.Dir}})
`

func mkdirCmd(dir, user string) (string, error) {
	// No-op if the directory already exists.
	t := template.Must(template.New("mkdir").Parse(mkdirCmdTmpl))
	d := struct {
		Dir  string
		User string
	}{
		Dir:  dir,
		User: user,
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, d); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Mkdir creates a directory in the console's file system.
func Mkdir(b *Burmilla, dir, user string) error {
	// Make sure /home/rancher/.ssh exists.
	mkdir, err := mkdirCmd(dir, user)
	if err != nil {
		return errcode.Annotate(err, "build mkdir script")
	}
	return execError(b.ExecRet([]string{
		"/bin/bash", "-c", mkdir,
	}))
}
