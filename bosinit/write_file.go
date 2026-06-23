package bosinit

import (
	"strconv"
)

// WriteFile specifies a file to be written onto the file system.
type WriteFile struct {
	Path        string
	Permissions string
	Owner       string
	Content     string
}

// FilePerm gererates file permission string for use in WriteFile.
func FilePerm(m int) string {
	return "0" + strconv.FormatInt(int64(m), 8)
}

// RCLocal creates cloud-init entry to add /etc/rc.local file on the target.
func RCLocal(content string) *WriteFile {
	return &WriteFile{
		Path:        "/etc/rc.local",
		Permissions: FilePerm(0744),
		Owner:       "root",
		Content:     content,
	}
}
