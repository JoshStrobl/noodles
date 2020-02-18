package main

import (
	"fmt"
	"github.com/JoshStrobl/trunk"
	"os"
	"path/filepath"
	"strings"
)

// This file contains our functionality for our Requires system

// RunRequires will run project pre/postrun function or a script before/after (based on), based on what is provided in requires
func RunRequires(operationType string, requires []string) {
	if len(requires) > 0 {
		trunk.LogInfo("Running Requires on " + operationType)

		for _, projectOrScriptName := range requires {
			scriptRunAfter := strings.HasSuffix(projectOrScriptName, ":after") // Determine if this should be only run after our project or main script, only applies to scripts
			projectOrScriptName = (strings.Split(projectOrScriptName, ":"))[0]

			if project, exists := noodles.Projects[projectOrScriptName]; exists { // If this project exists
				var plugin NoodlesPlugin

				if project.Plugin == "go" {
					plugin = &goPlugin
				} else if project.Plugin == "less" {
					plugin = &lessPlugin
				} else if project.Plugin == "typescript" {
					plugin = &typescriptPlugin
				} else {
					trunk.LogErr(fmt.Sprintf("Failed to get the plugin for project %s and type %s\n", projectOrScriptName, project.Plugin))
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

					if preRunErr := plugin.RequiresPreRun(&project); preRunErr != nil { // If we failed in our PreRun
						trunk.LogErr(fmt.Sprintf("Failed to run %s PreRun: %s\n", projectOrScriptName, preRunErr.Error()))
					}
				} else if operationType == "RequiresPostRun" { // If this is a PostRun operation
					if postRunErr := plugin.RequiresPostRun(&project); postRunErr != nil { // If we failed in our PostRun
						trunk.LogErr(fmt.Sprintf("Failed to run %s PostRun: %s\n", projectOrScriptName, postRunErr.Error()))
					}
				}
			} else if _, exists := noodles.Scripts[projectOrScriptName]; exists { // If this is a script
				if (operationType == "RequiresPreRun" && !scriptRunAfter) || // Running before
					(operationType == "RequiresPostRun" && scriptRunAfter) { // Running after and should run after
					RunScript(projectOrScriptName) // Call RunScript
				}
			}
		}
	}
}
