package main

// NoodlesConfig
// Configuration of Noodles and its projects
type NoodlesConfig struct {
	Name, Description, License, Version string
	Projects map[string]NoodlesProject
}

// NoodlesProject
// Configuration for Noodles Projects
type NoodlesProject struct {
	Destination, Mode, Plugin, Source string
	Flags, Requires []string
	Binary, Compress, Frala bool
}