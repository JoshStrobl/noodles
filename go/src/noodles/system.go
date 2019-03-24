package main

// This file contains functionality pertaining to interacting with the underlying package management systems of the host

import (
	"fmt"
	"github.com/stroblindustries/coreutils"
	"os/user"
	"runtime"
)

// CurrentUser is the current user running noodles
var CurrentUser string

// DependenciesMap is a map of plugins to their necessary dependencies
var DependenciesMap map[string]DependencyMap

// SystemPackager is the package manager of the host system (if any)
var SystemPackager string

func init() {
	user, _ := user.Current()
	CurrentUser = user.Username

	DependenciesMap = map[string]DependencyMap{ // Map of all the deps you'll need based on project configuration options
		"compress": { // Compression for TypeScript
			Binary:       "uglifyjs2",
			Dependencies: []string{"uglify-js2"},
			Packager:     "npm",
		},
		"go": { // Golang
			Binary:       "go",
			Dependencies: []string{"golang"},
			Packager:     "system",
		},
		"html": { // HTML
			Binary:       "html-minifier",
			Dependencies: []string{"html-minifier"}, // Uses html-minifier to minify HTML (ya don't say?)
			Packager:     "npm",
		},
		"less": {
			Binary:       "lessc",
			Dependencies: []string{"globby", "less", "less-plugin-clean-css", "less-plugin-glob"}, // LESS, Clean CSS plugin, Glob plugin
			Packager:     "npm",
		},
		"nodejs": { // nodejs (for dependencies requiring npm)
			Binary:       "npm",
			Dependencies: []string{"npm"},
			Packager:     "system",
		},
		"typescript": {
			Binary:       "tsc",
			Dependencies: []string{"typescript"}, // closurecompiler and Typescript are needed
			Packager:     "npm",
		},
	}

	SystemPackager = GetSystemPackageManager()

	if SystemPackager == "eopkg" { // If the package manager is eopkg (Solus)
		nodejsDepMap := DependenciesMap["nodejs"]
		nodejsDepMap.Dependencies = []string{"nodejs"} // npm is bundled with nodejs, thus require nodejs
		DependenciesMap["nodejs"] = nodejsDepMap
	}
}

// GetSystemPackageManager is responsible for returning the system's package manager, if any.
func GetSystemPackageManager() string {
	var pm string

	if runtime.GOOS == "linux" { // If we're running Linux
		packageManagers := []string{"apt", "dnf", "eopkg", "pacman", "zypper"}

		for _, bin := range packageManagers {
			if coreutils.ExecutableExists(bin) {
				pm = bin
				break
			}
		}

		if pm == "" { // If our package manager isn't set yet
			pm = "unknown" // If we've made it this far, we have no idea what this Linux user is using
		}
	} else { // If we're not running Linux, don't even try (sorry, I'd love PRs for more support)
		pm = "none"
	}

	return pm
}

// PackageInstaller will install the necessary packages / dependencies with the package manager or npm
func PackageInstaller(packager string, dependencies []string) {
	installFlags := []string{"install"} // Set the base "install" flag

	if packager == "system" { // Host package manager
		installFlags = append(installFlags, []string{"-y"}...) // Set to -y for auto-install
		packager = SystemPackager
	} else { // Nodejs
		installFlags = append(installFlags, []string{"--global"}...) // Globally install package
		packager = "npm"
	}

	installFlags = append(installFlags, dependencies...)                 // Append the dependencies / packages we'll be installing
	installOutput := coreutils.ExecCommand(packager, installFlags, true) // Call ExecCommand and output response to terminal
	fmt.Println(installOutput)
}
