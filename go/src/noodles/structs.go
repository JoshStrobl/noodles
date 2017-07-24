package main

// NoodlesConfig
// Configuration of Noodles and its projects
type NoodlesConfig struct {
	Description string "description"
	License string "license"
	Name string "name"
	Projects map[string]NoodlesProject "projects,flow"
	Version float64 "version"
}

// NoodlesProject
// Configuration for Noodles Projects
type NoodlesProject struct {
	Binary bool "binary,omitempty"
	Compress bool "compress,omitempty"
	Destination string "destination,omitempty"
	Flags []string "flags,omitempty"
	Frala bool "frala,omitempty"
	Mode string "mode,omitempty"
	Plugin string "plugin"
	Requires []string "requires,omitempty"
	Source string "source"
}