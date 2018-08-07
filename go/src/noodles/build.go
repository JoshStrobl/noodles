package main

// Build functionality

import (
	"fmt"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build all or a specific project",
	Long:  "Build all or a specific project",
	Run:   build,
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
		var plugin NoodlesPlugin

		switch project.Plugin {
		case "go": // Go
			plugin = &goPlugin
			break
		case "less": // LESS
			plugin = &lessPlugin
			break
		case "typescript": // TypeScript
			plugin = &typescriptPlugin
			break
		}

		fmt.Printf("Performing pre-run checks for %s\n", name)
		preRunErr := plugin.PreRun(&project)

		if preRunErr != nil { // If there was an error during pre-run
			fmt.Printf("An error occured during pre-run checks:\n%s\n", preRunErr.Error())
			return
		}

		fmt.Printf("Performing compilation for %s\n", name)
		runErr := plugin.Run(&project)

		if runErr != nil { // If there was an error during run
			fmt.Printf("An error occured during compilation:\n%s\n", runErr.Error())

			if project.Plugin != "go" { // If this isn't Go, where it's absolutely mandatory to do a GOPATH reset
				return
			}
		}

		fmt.Printf("Performing post-run for %s\n", name)
		postRunErr := plugin.PostRun(&project)

		if postRunErr != nil { // If there was an error during post-run
			fmt.Printf("An error occured during post-run:\n%s\n", postRunErr.Error())
		}
	} else {
		fmt.Println(name + " is not a valid project")
	}
}
