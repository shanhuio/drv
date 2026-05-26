package homeboot

import (
	"shanhu.io/std/docker"
)

const emptyDockerFile = `FROM scratch
MAINTAINER Shanhu Tech Inc.
CMD ["/bin/sleep", "1"]
`

// BuildEmpty builds the homedrv/empty image. This image is only used for
// processing volumes.
func BuildEmpty(client *docker.Client, name string) error {
	files := docker.NewTarStream(emptyDockerFile)
	return docker.BuildImageStream(client, name, files)
}

// BuildEmptyIfNotExist builds the homedrv/empty image if the image
// does not exist yet.
func BuildEmptyIfNotExist(client *docker.Client, name string) error {
	has, err := docker.HasImage(client, name)
	if err != nil {
		return err
	}
	if !has {
		return BuildEmpty(client, name)
	}
	return nil
}
