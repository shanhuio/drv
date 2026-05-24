package drvconfig

// NoServer to indicate there is no remote server.
const NoServer = "-"

// Config is the configuration of a HomeDrive. These configurations are
// critical for initializing an endpoint. The configurations here provides
// information about the drive's operating systems, network and identity,
// so that the drive can be properly installed and initialized.
type Config struct {
	Name string // Name of the endpoint.

	// Server address, default https://www.homedrive.io
	// "-" means no remote server is being used.
	Server string `json:",omitempty"`

	// Subscribe to which release channel.
	Channel string `json:",omitempty"`

	// Naming conventions. When this is null, using legacy naming.
	Naming *Naming `json:",omitempty"`

	// Identity PEM key file.
	IdentityPem string `json:",omitempty"`

	// Path to docker unix domain socket.
	DockerSock string `json:",omitempty"`

	// Path to system docker unix domain socket.
	SystemDockerSock string `json:",omitempty"`

	// Not running inside a docker.
	External bool `json:",omitempty"`

	// HTTPPort provides alternative http port for doorway container to
	// listen on. If it is negative, then doorway will not listen on
	// HTTP.
	HTTPPort int `json:",omitempty"`

	// HTTPSPort provides the alternative https port for doorway
	// container to listen on. If it is negative, then doorway will
	// not listen on HTTPS.
	HTTPSPort int `json:",omitempty"`

	// When this is true, do not bind ports when port number is 0 and it is not
	// managing the OS.
	AutoAvoidPortBinding bool `json:",omitempty"`

	// Instead of reading the endpoint init config from the server,
	// read from this file.
	EndpointInitConfigFile string `json:",omitempty"`
}
