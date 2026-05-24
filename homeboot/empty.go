package homeboot

import (
	"shanhu.io/g/dock"
)

const emptyDockerFile = `FROM scratch
MAINTAINER Shanhu Tech Inc.
CMD ["/bin/sleep", "1"]
`

// BuildEmpty builds the homedrv/empty image. This image is only used for
// processing volumes.
func BuildEmpty(client *dock.Client, name string) error {
	files := dock.NewTarStream(emptyDockerFile)
	return dock.BuildImageStream(client, name, files)
}

// BuildEmptyIfNotExist builds the homedrv/empty image if the image
// does not exist yet.
func BuildEmptyIfNotExist(client *dock.Client, name string) error {
	has, err := dock.HasImage(client, name)
	if err != nil {
		return err
	}
	if !has {
		return BuildEmpty(client, name)
	}
	return nil
}
