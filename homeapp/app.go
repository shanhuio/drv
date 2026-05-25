package homeapp

import (
	"shanhu.io/drv/drvapi"
)

// App is a generic application object that manages the lifecycle
// if an application running on a HomeDrive.
type App interface {
	// Called when the version is changed from a non-empty string to ver.
	// Normally the previous version would be a different version, but
	// in forced upgrades, it can also be the save version string.
	// change from a non-nil meta needs to stop() the service first.
	// change to a non-nil meta must auto start() the service.
	Change(from, to *drvapi.AppMeta) error

	// Send a soft signal to an app to start.
	Start() error

	// Send a soft signal to an app to stop.
	Stop() error
}
