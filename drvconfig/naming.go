package drvconfig

import (
	"path"
)

// Naming defines the naming conventions of a jarvis installation.
type Naming struct {
	// Network name. Default: "homedrv"
	Network string `json:",omitempty"`

	// Suffix for container and volume names. Default: ".homedrv"
	Suffix string `json:",omitempty"`

	// Image registry path of downloaded images. Default: "cr.homedrive.io"
	Registry string `json:",omitempty"`
}

// Default names
const (
	DefaultNetwork  = "homedrv"
	DefaultSuffix   = ".homedrv"
	DefaultRegistry = "cr.homedrive.io"
)

// Name returns the name of a container or volume.
func Name(n *Naming, cont string) string {
	if n == nil {
		return cont
	}
	suffix := n.Suffix
	if suffix == "" {
		suffix = DefaultSuffix
	}
	return cont + suffix
}

// Image returns the image name of an image type.
func Image(n *Naming, img string) string {
	reg := DefaultRegistry
	if n != nil && n.Registry != "" {
		reg = n.Registry
	}
	name := img
	if name == "jarvis" {
		name = "core"
	}
	const project = "homedrv"
	return path.Join(reg, project, name)
}

// Core returns the name of the core.
func Core(n *Naming) string {
	if n == nil {
		return "jarvis"
	}
	return Name(n, "core")
}

// OldCore returns the name of the old core.
func OldCore(n *Naming) string {
	if n == nil {
		return "jarvis-old"
	}
	return Name(n, "old.core")
}

// Network returns the network's name.
func Network(n *Naming) string {
	if n == nil || n.Network == "" {
		return DefaultNetwork
	}
	return n.Network
}
