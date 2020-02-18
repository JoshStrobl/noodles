package main

import (
	"fmt"
	"github.com/JoshStrobl/trunk"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strconv"
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
	if conf, readErr := ReadConfig(filepath.Join(workdir, "noodles.toml")); readErr == nil { // Read the config
		noodles = conf
	} else {
		fmt.Printf("noodles.toml appears to have the following issue(s):\n%s\n", readErr.Error())
		os.Exit(1)
	}

	for name, project := range noodles.Projects { // For each project
		var plugin NoodlesPlugin
		trunk.LogInfo("Checking " + name)

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
				trunk.LogFatal(project.Plugin + "is not a valid plugin.")
			}

			results := plugin.Check(&project) // Check the project, return our check results
			resultsTypes := []string{"Deprecations", "Errors", "Recommendations"}

			for _, resultType := range resultsTypes {
				if resultList, exists := results[resultType]; exists { // This type exists
					if len(resultList) != 0 { // There are items in this type
						header := fmt.Sprintf("%s (%s)", resultType, strconv.Itoa(len(resultList)))

						if resultType == "Deprecations" { // Deprecations
							trunk.LogWarn(header)
						} else if resultType == "Errors" { // Errors
							trunk.LogErr(header)
						} else if resultType == "Recommendations" { // Recommendations
							trunk.LogInfo(header)
						}

						for _, item := range resultList { // For each item
							fmt.Println(item)
						}
					} else {
						trunk.LogSuccess(fmt.Sprintf("%s: None", resultType))
					}
				} else {
					trunk.LogSuccess(fmt.Sprintf("%s: None", resultType))
				}
			}

			fmt.Println()
		} else if project.Plugin == "" && project.Requires == nil { // No Plugin or Requires are set
			trunk.LogErr(name + " is missing a plugin definition.")
		}
	}
}
