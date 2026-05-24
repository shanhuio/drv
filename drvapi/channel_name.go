package drvapi

import (
	"strings"
)

var archs = []string{"amd64", "arm64"}

func archSuffix(arch string) string {
	return "-" + arch
}

// ArchOf returns the architecture of
func ArchOf(name string) string {
	parsed := ParseChannelName(name)
	return parsed.Architecture()
}

// ChannelName is a parsed channel name.
type ChannelName struct {
	Base string
	Arch string
}

// Architecture returns the architecture of the channel.
func (n *ChannelName) Architecture() string {
	if n.Arch == "" {
		return "amd64"
	}
	return n.Arch
}

func (n *ChannelName) String() string {
	if n.Arch == "" {
		return n.Base
	}
	return n.Base + archSuffix(n.Arch)
}

// ParseChannelName parse the channel name into base name and architecture
func ParseChannelName(name string) *ChannelName {
	if name == "" {
		return nil
	}
	for _, arch := range archs {
		suffix := archSuffix(arch)
		if strings.HasSuffix(name, suffix) {
			return &ChannelName{
				Base: strings.TrimSuffix(name, suffix),
				Arch: arch,
			}
		}
	}
	return &ChannelName{Base: name}
}
