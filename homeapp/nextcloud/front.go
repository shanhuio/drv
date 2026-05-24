package nextcloud

import (
	"shanhu.io/g/dock"
	"shanhu.io/g/errcode"
	"shanhu.io/homedrv/drv/drvapi"
	drvcfg "shanhu.io/homedrv/drv/drvconfig"
	"shanhu.io/homedrv/drv/homeapp"
	"shanhu.io/homedrv/drv/homeapp/apputil"
)

// Front is the ncfront app.
type Front struct {
	core homeapp.Core
}

// NewFront creates a new ncfront app.
func NewFront(c homeapp.Core) *Front { return &Front{core: c} }

func (f *Front) cont() *dock.Cont {
	return dock.NewCont(f.core.Docker(), homeapp.Cont(f.core, NameFront))
}

func (f *Front) createCont(image string) (*dock.Cont, error) {
	if image == "" {
		return nil, errcode.InvalidArgf("no image specified")
	}

	nextcloudAddr := homeapp.Cont(f.core, Name) + ":80"
	config := &dock.ContConfig{
		Name:          homeapp.Cont(f.core, NameFront),
		Network:       homeapp.Network(f.core),
		Env:           map[string]string{"NEXTCLOUD": nextcloudAddr},
		AutoRestart:   true,
		JSONLogConfig: dock.LimitedJSONLog(),
		Labels:        drvcfg.NewNameLabel(NameFront),
	}
	return dock.CreateCont(f.core.Docker(), image, config)
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
