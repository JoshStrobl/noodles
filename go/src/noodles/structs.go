package main

// NoodlesConfig
// Configuration of Noodles and its projects
type NoodlesConfig struct {
	Description string
	License     string
	Name        string
	Projects    map[string]NoodlesProject
	Version     float64
}

// NoodlesProject
// Configuration for Noodles Projects
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
