package main

// NoodlesConfig
// Configuration of Noodles and its projects
type NoodlesConfig struct {
	Description string `toml: "description"`
	License string `toml: "license"`
	Name string `toml: "name"`
	Projects map[string]NoodlesProject `toml: "projects"`
	Version float64 `toml: "version"`
}

// NoodlesProject
// Configuration for Noodles Projects
type NoodlesProject struct {
	Binary bool `toml: "binary,omitempty"`
	Compress bool `toml: "compress,omitempty"`
	Destination string `toml: "destination,omitempty"`
	Flags []string `toml: "flags,omitempty"`
	Frala bool `toml: "frala,omitempty"`
	Mode string `toml: "mode,omitempty"`
	Plugin string `toml: "plugin"`
	Requires []string `toml: "requires,omitempty"`
	Source string `toml: "source"`
}