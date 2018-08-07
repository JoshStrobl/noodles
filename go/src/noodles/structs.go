package main

// NoodlesLintResult contains Deprecations, Errors, and Recommendations
type NoodlesLintResult map[string][]string

// NoodlesProject is the configuration for Noodles Projects.
type NoodlesProject struct {
	AppendHash      bool `toml:"AppendHash,omitempty"`
	Binary          bool `toml:"Binary,omitempty"`
	Compress        bool `toml:"Compress,omitempty"`
	Destination     string
	Flags           []string
	Mode            string `toml:"Mode,omitempty"`
	Plugin          string
	Requires        []string
	SimpleName      string `toml:"SimpleName,omitempty"`
	Source          string
	TarballLocation string `toml:"TarballLocation,omitempty"`
	Target          string `toml:"Target,omitempty"`
}

// NoodlesPlugin is an interface for plugins to implement
type NoodlesPlugin interface {
	// Lint is a function that will check the values of various aspects of a NoodlesProject and make recommendations
	Lint(n *NoodlesProject) NoodlesLintResult

	// PreRun is a function that should be performed prior to primary compilation
	PreRun(n *NoodlesProject) error

	// PostRun is a function that should be performed after primary compilation
	PostRun(n *NoodlesProject) error

	// Run is the primary compilation function
	Run(n *NoodlesProject) error
}

// NoodlesScript is the configuration for a Noodles Script
type NoodlesScript struct {
	Arguments   []string `toml:"Arguments,omitempty"`
	Description string   `toml:"Description,omitempty"`
	Directory   string   `toml:"Directory,omitempty"`
	Exec        string
	File        string `toml:"File,omitempty"`
	Redirect    bool   `toml:"Redirect,omitempty"`
}

type validateFunc func(string) error
