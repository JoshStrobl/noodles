package main

import (
	"fmt"
	"github.com/JoshStrobl/trunk"
	"github.com/spf13/cobra"
)

var tidyCmd = &cobra.Command{
	Use:               "tidy",
	Short:             "Runs available tidying utilities for projects",
	Long:              "Runs available tidying utilities for projects",
	Run:               tidy,
	DisableAutoGenTag: true,
}

var tidyProject string

func init() {
	tidyCmd.Flags().StringVarP(&tidyProject, "project", "p", "", "Name of the project we're tidying")
}

func tidy(cmd *cobra.Command, args []string) {
	if tidyProject == "" { // If no project is set
		for name := range noodles.Projects { // For each project
			TidyProject(name)
		}
	} else { // If a specific project is set
		TidyProject(tidyProject)
	}
}

// TidyProject is responsible for running the respective tidying functions for each project's type
func TidyProject(name string) {
	if project, exists := noodles.Projects[name]; exists { // If this project exists
		if project.Plugin == "go" { // Go plugin
			plugin := &goPlugin

			trunk.LogInfo("Performing pre-run checks for " + name)
			preRunErr := plugin.PreRun(&project)

			if preRunErr != nil { // If there was an error during pre-run
				trunk.LogErrRaw(fmt.Errorf("An error occurred during pre-run checks:\n%s\n", preRunErr.Error()))
				return
			}

			plugin.ModTidy(&project) // Tidy up Go Modules

			trunk.LogInfo("Performing post-run for " + name)
			postRunErr := plugin.PostRun(&project)

			if postRunErr != nil { // If there was an error during post-run
				trunk.LogErrRaw(fmt.Errorf("An error occurred during post-run:\n%s\n", postRunErr.Error()))
			}
		}
	} else {
		trunk.LogErr(name + " is not a valid project")
	}
}
