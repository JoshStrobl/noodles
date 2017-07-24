package main

// NoodlesConfig
// Configuration of Noodles and its projects
type NoodlesConfig struct {
	Name string "name"
	Description string "description"
	License string "license"
	Version float64 "version"
	Projects map[string]NoodlesProject "projects,flow"
}

// NoodlesProject
// Configuration for Noodles Projects
type NoodlesProject struct {
	Destination string "destination,omitempty"
	Mode string "mode,omitempty"
	Plugin string "plugin"
	Source string "source"
	Flags []string "flags,omitempty"
	Requires []string "flags,omitempty"
	Binary bool "binary,omitempty"
	Compress bool "compress,omitempty"
	Frala bool "frala,omitempty"
}