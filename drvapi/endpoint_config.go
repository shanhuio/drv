package drvapi

// EndpointInitConfig is the basic settings to initialize an endpoint. This only
// affects the init time of an endpoint. After an endpoint is provisioned, a
// user might be able to change the configuration via jarvis' user interface.
type EndpointInitConfig struct {
	// Main domain. Will serve jarvis Web UI. Currently redirects to the first
	// Nextcloud domain. If missing, an endpoint's main domain is
	// <name>.homedrv.com
	MainDomain string `json:",omitempty"`

	// Apps is the list of apps to install on initialization.
	Apps []string `json:",omitempty"`

	// Nextcloud domains. Incoming traffic for these domains will be redirected
	// to nextcloud. If empty, will use nextcloud.<name>.homedrv.com.
	NextcloudDomains []string `json:",omitempty"`

	// Extra domains that will be routed to this endpoint. Using those needs
	// custom doorway host map settings.
	ExtraDomains []string `json:",omitempty"`

	// Fabrics server to connect to. Default using "fabrics.homedrive.io"
	//
	// TODO(h8liu): deprecate this and get the shareded fabrics server
	// address from the centralized server.
	FabricsServer string `json:",omitempty"`
}
