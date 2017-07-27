package main

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
	Binary      bool
	Compress    bool
	Destination string
	Flags       []string
	Frala       bool
	Mode        string
	Plugin      string
	Requires    []string
	Source      string
}
