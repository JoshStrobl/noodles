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

func build(cmd *cobra.Command, args []string) {
	if project == "" { // If no project is set
		for name := range noodles.Projects { // For each project
			BuildProject(name)
		}
	} else { // If a specific project is set
		BuildProject(project)
	}
}

// BuildProject is responsible for determining the appropriate plugin to execute and handle requires.
func BuildProject(name string) {
	if project, exists := noodles.Projects[name]; exists { // If this project exists
		fmt.Println("Building " + name)
		switch project.Plugin {
		case "go": // Go
			project.Go(name) // Run the Go plugin
		case "typescript": // TypeScript
			project.Typescript(name) // Run the TypeScript plugin
		default: // Not a valid name
			fmt.Println("Invalid plugin.")
		}
	} else {
		fmt.Println(name + " is not a valid project")
	}
}
