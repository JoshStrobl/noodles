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

			resultsTypes := []string { "Deprecations", "Errors", "Recommendation" }

			fmt.Printf("Linting %s:\n", name)

			for _, resultType := range resultsTypes {
				if resultList, exists := results[resultType]; exists { // This type exists
					if len(resultList) != 0 { // There are items in this type
						fmt.Printf("\t%s:\n", resultType)
						for _, item := range resultList { // For each item
							fmt.Printf("\t\t❌ %s\n", item)
						}
					} else {
						fmt.Printf("\t%s: None\n", resultType)
					}
				} else {
					fmt.Printf("\t%s: None\n", resultType)
				}
			}

			fmt.Println()
		} else if project.Plugin == "" && project.Requires == nil { // No Plugin or Requires are set
			fmt.Printf("%s is missing a plugin definition.\n", name)
		}
	}
}