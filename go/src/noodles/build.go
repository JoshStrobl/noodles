package main

// Build functionality

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var buildCmd = &cobra.Command{
	Use:               "build",
	Short:             "Build all or a specific project",
	Long:              "Build all or a specific project",
	Run:               build,
	DisableAutoGenTag: true,
}

var buildProject string

func init() {
	buildCmd.Flags().StringVarP(&buildProject, "project", "p", "", "Name of a project we're building")
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
		RunRequiresOperation("RequiresPreRun", &project)

		var plugin NoodlesPlugin

		if project.Plugin == "go" {
			plugin = &goPlugin
		} else if project.Plugin == "less" {
			plugin = &lessPlugin
		} else if project.Plugin == "typescript" {
			plugin = &typescriptPlugin
		} else {
			fmt.Printf("Failed to get the plugin for type: %s\n", project.Plugin)
			return
		}

		fmt.Printf("Performing pre-run checks for %s\n", name)
		preRunErr := plugin.PreRun(&project)

		if preRunErr != nil { // If there was an error during pre-run
			fmt.Printf("An error occurred during pre-run checks:\n%s\n", preRunErr.Error())
			return
		}

		fmt.Printf("Performing compilation for %s\n", name)
		runErr := plugin.Run(&project)

		if runErr != nil { // If there was an error during run
			fmt.Printf("An error occurred during compilation:\n%s\n", runErr.Error())

			if project.Plugin != "go" { // If this isn't Go, where it's absolutely mandatory to do a GOPATH reset
				return
			}
		}

		RunRequiresOperation("RequiresPostRun", &project)

		fmt.Printf("Performing post-run for %s\n", name)
		postRunErr := plugin.PostRun(&project)

		if postRunErr != nil { // If there was an error during post-run
			fmt.Printf("An error occurred during post-run:\n%s\n", postRunErr.Error())
		}
	} else {
		fmt.Println(name + " is not a valid project")
	}
}

// RunRequiresOperation will run an operation (pre-run or post-run require funcs) against the specified project's Requires
func RunRequiresOperation(operationType string, n *NoodlesProject) {
	if len(n.Requires) > 0 {
		fmt.Printf("Running %s on project's Requires.\n", operationType)

		for _, requiredProjectName := range n.Requires { // For each required project
			if project, exists := noodles.Projects[requiredProjectName]; exists { // If this project exists
				var plugin NoodlesPlugin

				if project.Plugin == "go" {
					plugin = &goPlugin
				} else if project.Plugin == "less" {
					plugin = &lessPlugin
				} else if project.Plugin == "typescript" {
					plugin = &typescriptPlugin
				} else {
					fmt.Printf("Failed to get the plugin for project %s and type %s\n", requiredProjectName, project.Plugin)
					return
				}

				if operationType == "RequiresPreRun" { // If this is a PreRun operation
					if project.Plugin == "go" { // If this is a Go-type project
						if currentWd, getWdErr := os.Getwd(); getWdErr == nil { // Get the current working directory
							if filepath.Base(currentWd) != "go" { // Currently not in go directory
								os.Chdir(filepath.Join(workdir, "go")) // Change to our Go directory
							}
						}
					}

					plugin.RequiresPreRun(&project)
				} else if operationType == "RequiresPostRun" { // If this is a PostRun operation
					plugin.RequiresPostRun(&project)
				}
			} else {
				fmt.Printf("Project %s does not exist.\n", requiredProjectName)
				return
			}
		}
	}
}
