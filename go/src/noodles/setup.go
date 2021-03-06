package main

import (
	"fmt"
	"github.com/JoshStrobl/trunk"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"strings"
)

// DependencyMap describes the dependencies you'll need and whether you need them from the system package manager or a separate packaging system
type DependencyMap struct {
	Binary       string
	Dependencies []string
	Packager     string
}

var setupCmd = &cobra.Command{
	Use:               "setup",
	Short:             "Set up all or a specific project",
	Long:              "Set up all or a specific project. This will attempt to install the necessary dependencies required for various projects.",
	Run:               Setup,
	DisableAutoGenTag: true,
}

// ExecutableMissing is the message when a specific executable is missing.
const ExecutableMissing = "%s does not exist on this system."

// PluginNowSetup is the message when a plugin is now set up.
const PluginNowSetup = "%s is now set up."

// PluginAlreadySetup is the message when a plugin has already been set up.
const PluginAlreadySetup = "%s is already set up."

var setupProject string

func init() {
	setupCmd.Flags().StringVarP(&setupProject, "project", "p", "", "Name of a project we're setting up")
}

// Setup will set up all or a specific project, checking for various binaries, dependencies, and attempt to install requirements.
func Setup(cmd *cobra.Command, args []string) {
	var projects map[string]NoodlesProject

	if setupProject != "" { // If a project has been specified
		if projectInfo, exists := noodles.Projects[setupProject]; exists {
			projects = map[string]NoodlesProject{
				setupProject: projectInfo,
			}
		} else { // If project does not exist
			trunk.LogFatal(setupProject + " is not a valid project.")
		}
	} else { // If no project has been specified
		projects = noodles.Projects
	}

	for name, projectInfo := range projects { // For each project
		dependenciesExist, dependenciesMissing := HasDependencies(projectInfo) // Check if we have dependencies for this plugin

		if dependenciesExist {
			trunk.LogSuccess(fmt.Sprintf("Dependencies for %s are already satisfied.", name))
		} else {
			trunk.LogInfo(fmt.Sprintf("Dependencies for %s are not satisfied.", name))
			PrintSummary(dependenciesMissing)

			if SystemPackager != "unknown" && SystemPackager != "none" { // If we've determined a valid package manager
				if CurrentUser == "root" { // If we are running as root
					if projectInfo.Plugin == "go" { // If this is Go
						PackageInstaller("system", []string{"go"})
					} else { // If this is an NPM-based set of packages, or nodejs / npm itself
						firstItem := dependenciesMissing[0]

						if firstItem == "nodejs" { // If the first item is nodejs
							PackageInstaller("system", []string{"nodejs"}) // Install nodejs before installing any NPM packages
							dependenciesMissing = dependenciesMissing[1:]  // Update dependenciesMissing to not include nodejs
						}

						PackageInstaller("npm", dependenciesMissing) // Now install NPM packages
					}
				} else { // If we are not running as root
					trunk.LogFatal("You must run noodles setup with sudo (root) to install the necessary dependencies.")
				}
			} else {
				trunk.LogFatal("Unable to determine the appropriate package manager, if any. Please manually install dependencies.")
			}
		}
	}
}

// HasDependencies will check for the necessary dependencies for a plugin
func HasDependencies(p NoodlesProject) (bool, []string) {
	var depsExist bool
	var depsMissing []string

	if p.Plugin != "go" { // If the plugin isn't Go
		pluginDepMap := DependenciesMap[p.Plugin] // Get the dependency map for this plugin
		binaries := []string{pluginDepMap.Binary} // Set binaries to a slice of strings, where our initial string is our primary binary

		if p.Plugin == "typescript" && p.Compress { // If the project uses the Typescript plugin as well as compression (needs terser)
			binaries = append(binaries, DependenciesMap["compress"].Binary) // Include our compression binary as well
		}

		if depsExist = coreutils.ExecutableExists("npm"); !depsExist { // If npm exists
			depsMissing = []string{"nodejs"}
		}

		for _, binary := range binaries { // For each binary
			if depsExist = coreutils.ExecutableExists(binary); !depsExist { // If the binary does not exist
				if binary == "terser" { // If this is terser
					pluginDepMap = DependenciesMap["compress"] // Change pluginDepMap to the one for compress
				}

				depsMissing = append(depsMissing, pluginDepMap.Dependencies...) // Add the npm packages we may be missing
			}
		}
	} else { // If plugin is Go
		if depsExist = coreutils.ExecutableExists("go"); !depsExist { // If go does not exist
			depsMissing = []string{"go"}
		}
	}

	return depsExist, depsMissing
}

// PrintSummary will print a summary of missing dependencies
func PrintSummary(missing []string) {
	firstItem := missing[0]
	var npmPackages []string

	if firstItem == "go" || firstItem == "nodejs" { // If we're missing a system-level dependency
		systemDepMap := DependenciesMap[firstItem] // Get the dependency map for this system package

		trunk.LogInfo("Missing System Package: " + systemDepMap.Dependencies[0])
		npmPackages = missing[1:] // Set any extra dependencies to npmPackages
	} else { // If the first item is not a system-level package
		npmPackages = missing // Set all of missing to npmPackages
	}

	if len(npmPackages) != 0 { // If we're missing NPM packages
		trunk.LogInfo(fmt.Sprintf("Missing NPM Package(s): %s", strings.Join(npmPackages, ", ")))
	}
}
