package drvapi

// StepVersion records a particular version of a nextcloud or postgres release.
type StepVersion struct {
	Major   int    // Major version.
	Version string // Full version string.

	// Source is where this image originally come from. Information only.
	Source string `json:",omitempty"`

	// Image ID, often the hash of the image.
	Image string

	// ImageSum is the checksum of the image's gzipped tarball.
	ImageSum string `json:",omitempty"`
}

// AppMeta stores the meta information of an HomeDrive application
type AppMeta struct {
	Name string

	// Dependencies.
	Deps []string `json:",omitempty"`

	// Version counter. To prevent rolling back. Most apps
	// do not support rolling back.
	Version int64 `json:",omitempty"`

	// SemVersion tracks compability.
	SemVersion string `json:",omitempty"`

	// Image ID, for simple single container apps.
	Image string `json:",omitempty"`

	// ImageSum is the checksum of the image's gzipped tarball.
	ImageSum string `json:",omitempty"`

	// Steps is for apps that needs an upgrade ladder.
	Steps []*StepVersion `json:",omitempty"`
}
