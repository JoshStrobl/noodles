package main

// Build functionality

import (
	"fmt"
	"github.com/JoshStrobl/trunk"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:               "build",
	Short:             "Build all or a specific project",
	Long:              "Build all or a specific project",
	Run:               build,
	DisableAutoGenTag: true,
}

var buildProject string
var debug bool

func init() {
	buildCmd.Flags().StringVarP(&buildProject, "project", "p", "", "Name of a project we're building")
	buildCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable Debug Mode")
}

func build(cmd *cobra.Command, args []string) {
	if buildProject == "" { // If no project is set
		for name := range noodles.Projects { // For each project
			BuildProject(name)
		}
	} else { // If a specific project is set
		BuildProject(buildProject)
	}
}

// BuildProject is responsible for determining the appropriate plugin to execute and handle requires.
func BuildProject(name string) {
	if project, exists := noodles.Projects[name]; exists { // If this project exists
		RunRequires("RequiresPreRun", project.Requires)

		var plugin NoodlesPlugin

		if project.Plugin == "go" {
			plugin = &goPlugin
		} else if project.Plugin == "less" {
			plugin = &lessPlugin
		} else if project.Plugin == "typescript" {
			plugin = &typescriptPlugin
		} else {
			trunk.LogErr("Failed to get the plugin for type: " + project.Plugin)
			return
		}

		trunk.LogInfo("Performing pre-run checks for " + name)
		preRunErr := plugin.PreRun(&project)

		if preRunErr != nil { // If there was an error during pre-run
			trunk.LogErrRaw(fmt.Errorf("An error occurred during pre-run checks:\n%s\n", preRunErr.Error()))
			return
		}

		trunk.LogInfo("Performing compilation for " + name)
		runErr := plugin.Run(&project)

		if runErr == nil { // Compilation was successful
			trunk.LogSuccess(fmt.Sprintf("Built %s", name))
		} else {
			trunk.LogErrRaw(fmt.Errorf("An error occurred during compilation:\n%s\n", runErr.Error()))

			if project.Plugin != "go" { // If this isn't Go, where it's absolutely mandatory to do a GOPATH reset
				return
			}
		}

		RunRequires("RequiresPostRun", project.Requires)

		trunk.LogInfo("Performing post-run for " + name)
		postRunErr := plugin.PostRun(&project)

		if postRunErr != nil { // If there was an error during post-run
			trunk.LogErrRaw(fmt.Errorf("An error occurred during post-run:\n%s\n", postRunErr.Error()))
		}
	} else {
		trunk.LogErr(name + " is not a valid project")
	}
}
