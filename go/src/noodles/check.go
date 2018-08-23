package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var checkCmd = &cobra.Command{
	Use:               "check",
	Aliases:           []string{"validate"},
	Short:             "Validates the existing noodles.toml",
	Long:              "Validates the existing noodles.toml",
	Run:               check,
	DisableAutoGenTag: true,
}

// check will validate noodles.toml
func check(cmd *cobra.Command, args []string) {
	readErr := ReadConfig() // Read the config

	if readErr != nil { // If there was a read error on the config
		fmt.Printf("noodles.toml appears to have the following issue(s):\n%s\n", readErr.Error())
		os.Exit(1)
	}

	for name, project := range noodles.Projects { // For each project
		var plugin NoodlesPlugin
		fmt.Printf("Checking %s:\n", name)

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
			default:
				fmt.Printf("%s is not a valid plugin.\n", project.Plugin)
				os.Exit(1)
			}

			results := plugin.Check(&project) // Check the project, return our check results
			resultsTypes := []string{"Deprecations", "Errors", "Recommendation"}

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
