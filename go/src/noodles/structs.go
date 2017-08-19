package main

// DependencyMap
// Describes the dependencies you'll need and whether you need them from the system package manager or a separate packaging system
type DependencyMap struct {
	Binary string
	Dependencies []string
	Packager     string
}

// NoodlesConfig is the configuration of global properties of Noodles.
type NoodlesConfig struct {
	Description string
	License     string
	Name        string
	Projects    map[string]NoodlesProject
	Version     float64
}

// NoodlesProject is the configuration for Noodles Projects.
type NoodlesProject struct {
	Binary      bool `toml:"Binary,omitempty"`
	Compress    bool `toml:"Compress,omitempty"`
	Destination string
	Flags       []string
	Frala       bool `toml:"Frala,omitempty"`
	Mode        string `toml:"Mode,omitempty"`
	Plugin      string
	Requires    []string
	Source      string
	Target string `toml:"Target,omitempty"`
}
