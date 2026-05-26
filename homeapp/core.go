package homeapp

import (
	drvcfg "shanhu.io/drv/drvconfig"
	"shanhu.io/g/settings"
	"shanhu.io/std/docker"
)

// Core provides the core interface to run an application.
type Core interface {
	// App gets an application by name.
	App(name string) (App, error)

	// Docker gets the client to the application docker.
	Docker() *docker.Client

	// Settings gets the settings table.
	Settings() settings.Settings

	// Naming gets the naming convention of the drive. We might want to
	// migrate the legacy stuff and deprecate this some day.
	Naming() *drvcfg.Naming

	// Domains gets the stub that manages application domain routings.
	Domains() Domains
}

// Cont returns the container name of an app.
func Cont(c Core, app string) string { return drvcfg.Name(c.Naming(), app) }

// Network returns the network name.
func Network(c Core) string { return drvcfg.Network(c.Naming()) }

// Vol returns the volume name of an app.
func Vol(c Core, app string) string { return drvcfg.Name(c.Naming(), app) }
