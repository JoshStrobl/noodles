package main

// NoodlesConfig
// Configuration of Noodles and its projects
type NoodlesConfig struct {
	name, description, license, version string
	projects map[string]NoodlesProject
}

// NoodlesProject
// Configuration for Noodles Projects
type NoodlesProject struct {
	destination, mode, plugin, source string
	flags, requires []string
	binary, compress, frala bool
}