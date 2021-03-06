package main

// NoodlesCheckResult contains Deprecations, Errors, and Recommendations
type NoodlesCheckResult map[string][]string

// NoodlesProject is the configuration for Noodles Projects.
type NoodlesProject struct {
	AppendHash               bool `toml:"AppendHash,omitempty"`
	Compress                 bool `toml:"Compress,omitempty"`
	ConsolidateChildDirs     bool `toml:ConsolidateChildDirs,omitempty"`
	Destination              string
	DisableNestedEnvironment bool     `toml:"DisableNestedEnvironment,omitempty"`
	EnableGoModules          bool     `toml:"EnableGoModules,omitempty"`
	ExcludeItems             []string `toml:"ExcludeItems,omitempty"`
	Flags                    []string
	Mode                     string `toml:"Mode,omitempty"`
	Plugin                   string
	Private                  []string `toml:"Private,omitempty"`
	Requires                 []string
	SimpleName               string `toml:"SimpleName,omitempty"`
	Source                   string
	SourceDir                string `toml:"-"`
	TarballLocation          string `toml:"TarballLocation,omitempty"`
	Target                   string `toml:"Target,omitempty"`
	Type                     string `toml:"Type,omitempty"`
}

// NoodlesPlugin is an interface for plugins to implement
type NoodlesPlugin interface {
	// Check is a function that will check the values of various aspects of a NoodlesProject and make recommendations
	Check(n *NoodlesProject) NoodlesCheckResult

	// Lint is a function that will run any respective linting suits for a NoodlesPlugin against a NoodlesProject
	Lint(n *NoodlesProject, confidence float64) error

	// PreRun is a function that should be performed prior to primary compilation
	PreRun(n *NoodlesProject) error

	// PostRun is a function that should be performed after primary compilation
	PostRun(n *NoodlesProject) error

	// RequiresPreRun is a function that should be performed before PreRun, should the project be required by another
	RequiresPreRun(n *NoodlesProject) error

	// RequiresPostRun is a function that should be performed before PostRun, should the project be required by another
	RequiresPostRun(n *NoodlesProject) error

	// Run is the primary compilation function
	Run(n *NoodlesProject) error
}

// NoodlesScript is the configuration for a Noodles Script
type NoodlesScript struct {
	Arguments   []string `toml:"Arguments,omitempty"`
	Description string   `toml:"Description,omitempty"`
	Directory   string   `toml:"Directory,omitempty"`
	Exec        string
	File        string   `toml:"File,omitempty"`
	Redirect    bool     `toml:"Redirect,omitempty"`
	Requires    []string `toml:"Requires,omitempty"`
	UseGoEnv    bool     `toml:"UseGoEnv,omitempty"`
}

type validateFunc func(string) error
