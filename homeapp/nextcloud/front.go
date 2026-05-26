package nextcloud

import (
	"shanhu.io/drv/drvapi"
	drvcfg "shanhu.io/drv/drvconfig"
	"shanhu.io/drv/homeapp"
	"shanhu.io/drv/homeapp/apputil"
	"shanhu.io/std/docker"
	"shanhu.io/std/errcode"
)

// Front is the ncfront app.
type Front struct {
	core homeapp.Core
}

// NewFront creates a new ncfront app.
func NewFront(c homeapp.Core) *Front { return &Front{core: c} }

func (f *Front) cont() *docker.Cont {
	return docker.NewCont(f.core.Docker(), homeapp.Cont(f.core, NameFront))
}

func (f *Front) createCont(image string) (*docker.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}

	nextcloudAddr := homeapp.Cont(f.core, Name) + ":80"
	config := &docker.ContConfig{
		Name:          homeapp.Cont(f.core, NameFront),
		Network:       homeapp.Network(f.core),
		Env:           map[string]string{"NEXTCLOUD": nextcloudAddr},
		AutoRestart:   true,
		JSONLogConfig: docker.LimitedJSONLog(),
		Labels:        drvcfg.NewNameLabel(NameFront),
	}
	return docker.CreateCont(f.core.Docker(), image, config)
}

func (f *Front) startWithImage(image string) error {
	cont, err := f.createCont(image)
	if err != nil {
		return errcode.Annotate(err, "create ncfront container")
	}
	return cont.Start()
}

func (f *Front) install(image string) error {
	return f.startWithImage(image)
}

// Start starts the app.
func (f *Front) Start() error { return f.cont().Start() }

// Stop stops the app.
func (f *Front) Stop() error { return f.cont().Stop() }

// Change changes the app's version.
func (f *Front) Change(from, to *drvapi.AppMeta) error {
	if from != nil {
		if err := apputil.DropIfExists(f.cont()); err != nil {
			return errcode.Annotate(err, "drop old ncfront container")
		}
	}
	if to == nil {
		return nil
	}
	return f.install(homeapp.Image(to))
}
