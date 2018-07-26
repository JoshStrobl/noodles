package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Validates the existing noodles.toml",
	Long:  "Validates the existing noodles.toml",
	Run:   lint,
}

// lint will validate noodles.toml
func lint(cmd *cobra.Command, args []string) {
	readErr := ReadConfig() // Read the config

	if readErr != nil { // If there was a read error on the config
		fmt.Printf("noodles.toml appears to have the following issue(s):\n%s\n", readErr.Error())
		os.Exit(1)
	}

	for name, project := range noodles.Projects { // For each project
		var plugin NoodlesPlugin

		if project.Plugin != "" {
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

			results := plugin.Lint(&project) // Lint the project, return our lint results

			if (len(results.Deprecations) == 0) && (len(results.Errors) == 0) && (len(results.Recommendations) == 0) { // No issues at all
				fmt.Printf("%s has no issues.\n", name)
			} else {
				fmt.Printf("We found the following items that should be resolved with %s\n", name)

				if len(results.Deprecations) != 0 { // There are deprecations
					fmt.Println("\tDeprecations:")
					for _, deprecation := range results.Deprecations { // For each deprecation
						fmt.Printf("\t\t%s\n", deprecation)
					}
				}

				if len(results.Errors) != 0 { // There are errors
					fmt.Println("\tErrors:")
					for _, err := range results.Errors {
						fmt.Printf("\t\t%s\n", err)
					}
				}

				if len(results.Recommendations) != 0 { // There are recommendations
					fmt.Println("\tRecommendations:")
					for _, recommendation := range results.Recommendations {
						fmt.Printf("\t\t%s\n", recommendation)
					}
				}
			}
		} else if project.Plugin == "" && project.Requires == nil { // No Plugin or Requires are set
			fmt.Printf("%s is missing a plugin definition.\n", name)
		}
	}
}
