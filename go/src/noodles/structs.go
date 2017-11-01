package main

// DependencyMap
// Describes the dependencies you'll need and whether you need them from the system package manager or a separate packaging system
type DependencyMap struct {
	Binary       string
	Dependencies []string
	Packager     string
}

// NoodlesConfig is the configuration of global properties of Noodles.
type NoodlesConfig struct {
	Description string
	License     string
	Name        string
	Projects    map[string]NoodlesProject
	Scripts     map[string]NoodlesScript
	Version     float64
}

// NoodlesProject is the configuration for Noodles Projects.
type NoodlesProject struct {
	Binary          bool `toml:"Binary,omitempty"`
	Compress        bool `toml:"Compress,omitempty"`
	Destination     string
	Flags           []string
	Frala           bool   `toml:"Frala,omitempty"`
	Mode            string `toml:"Mode,omitempty"`
	Plugin          string
	Requires        []string
	Source          string
	TarballLocation string `toml:"TarballLocation,omitempty"`
	Target          string `toml:"Target,omitempty"`
}

// NoodlesScript is the configuration for a Noodles Script
type NoodlesScript struct {
	Arguments   []string `toml:"Arguments,omitempty"`
	Description string   `toml:"Description,omitempty"`
	Directory   string   `toml:"Directory,omitempty"`
	Exec        string
}
