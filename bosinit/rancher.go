package bosinit

// Rancher contains configuration for Rancher/Burmilla OS.
type Rancher struct {
	SSH          *SSH       `yaml:",omitempty"`
	Upgrade      *Upgrade   `yaml:",omitempty"`
	CloudInit    *CloudInit `yaml:"cloud_init,omitempty"`
	ResizeDevice string     `yaml:"resize_device,omitempty"`
}

// SSH has configurations for the Rancher/Burmilla OS SSH service.
type SSH struct {
	Port int
}

// Upgrade sets the upgrade source.
type Upgrade struct {
	URL string
}

// CloudInit specifies the cloud init properties.
type CloudInit struct {
	DataSources []string `yaml:"datasources,omitempty"`
}
