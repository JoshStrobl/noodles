package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var lintCmd = &cobra.Command{
	Use:               "lint",
	Short:             "Runs available linters for projects",
	Long:              "Runs available linters for projects",
	Run:               lint,
	DisableAutoGenTag: true,
}

var minimumConfidence float64
var lintProject string

func init() {
	lintCmd.Flags().Float64VarP(&minimumConfidence, "confidence", "c", 0.5, "Minimum confidence for linting problems")
	lintCmd.Flags().StringVarP(&lintProject, "project", "p", "", "Name of the project we're linting")
}

func lint(cmd *cobra.Command, args []string) {
	if lintProject == "" { // If no project is set
		for name := range noodles.Projects { // For each project
			LintProject(name)
		}
	} else { // If a specific project is set
		LintProject(buildProject)
	}
}

// LintProject is responsible for running the respective linters for each project's type
func LintProject(name string) {
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
			fmt.Printf("An error occurred during pre-run checks:\n%s\n", preRunErr.Error())
			return
		}

		lintErr := plugin.Lint(&project, minimumConfidence)

		if lintErr != nil {
			fmt.Printf("An error occurred during linting:\n%s\n", lintErr.Error())

			if project.Plugin != "go" { // If this isn't Go, where it's absolutely mandatory to do a GOPATH reset
				return
			}
		}

		if project.Plugin == "go" { // Only post-run is required for Go to reset env
			fmt.Printf("Performing post-run for %s\n", name)
			postRunErr := plugin.PostRun(&project)

			if postRunErr != nil { // If there was an error during post-run
				fmt.Printf("An error occurred during post-run:\n%s\n", postRunErr.Error())
			}
		}
	} else {
		fmt.Println(name + " is not a valid project")
	}
}
