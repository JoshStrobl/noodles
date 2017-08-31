package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"path/filepath"
	"regexp"
	"strings"
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

	if readErr == nil {
		configIsCorrect := true // Assume the TOML is correct

		for name, project := range noodles.Projects { // For each project
			if project.Plugin != "" {
				switch project.Plugin {
				case "go":
					if !strings.HasSuffix(project.Source, "*.go") { // Globbing isn't enabled
						configIsCorrect = false
						sourcePath := filepath.Dir(project.Source)
						fmt.Println(name + " is not using globbing for getting all Go files.")
						fmt.Printf("Recommend changing %s to %s\n", project.Source, filepath.Join(sourcePath, "*.go"))
					}
				case "typescript":
					if !project.Compress { // Compression not enabled
						configIsCorrect = false
						fmt.Println("Compression is not set for %s, meaning we will only generate a non-minified JS file. Recommended enabling Compress.\n", name)
					}

					if project.Mode == "" { // No mode is set
						configIsCorrect = false
						fmt.Printf("No mode is set for %s, meaning we will default to our Advanced flag set. Recommend setting a Mode to simple, advanced, or strict.\n", name)
					} else if (project.Mode != "simple") && (project.Mode != "advanced") && (project.Mode != "strict") { // Not a valid mode
						configIsCorrect = false
						fmt.Printf("%s is not a valid Mode for %s. Please change it to simple, advanced, or strict.\n", project.Mode, name)
					}

					if project.Target == "" { // No target set
						configIsCorrect = false
						fmt.Printf("No target is set for %s, meaning we will default to ES5. Recommended setting Target to ES5, ES6, or ES7.\n", name)
					} else if project.Target == "ES3" { // We're opinionated.
						configIsCorrect = false
						fmt.Printf("ES3 is not considered a valid target. Please change %s to ES5, ES6, or ES7.\n", name)
					} else if (project.Target != "ES5") && (project.Target != "ES6") && (project.Target != "ES7") { // Not a valid target
						configIsCorrect = false
						fmt.Printf("%s is not a valid target for %s. Please change it to ES5, ES6, or ES7.\n", project.Target, name)
					}
				}
			} else if project.Plugin == "" && project.Requires == nil { // No Plugin or Requires are set
				configIsCorrect = false
				fmt.Printf("%s is missing a plugin definition.\n", name)
			}
		}

		if configIsCorrect {
			fmt.Println("noodles.toml appears correct.")
		}

		SaveConfig() // Save to ensure we enforce indentation
	} else {
		cleanMessage := CleanLintErrors(readErr.Error())
		fmt.Printf("noodles.toml appears to have the following issue(s):\n%s\n", cleanMessage)
	}
}

// CleanLintErrors will remove some verbosity from any unmarshalling error
func CleanLintErrors(err string) string {
	if strings.HasPrefix(err, "noodles") { // If this is already a custom error message
		return err
	}

	cleanMessage := strings.Replace(err, "unmarshal", "convert", -1) // Change "unmarshal" to a human language
	cleanMessage = strings.Replace(cleanMessage, "!!", "", -1)
	cleanMessage = strings.Replace(cleanMessage, "`", "", -1) // Remove any ` wrapping types

	re := regexp.MustCompile(`line\s\d+:\s[\s\S]+$`) // Only get line N: message
	lineErrors := re.FindAllString(cleanMessage, -1) // Find all strings
	cleanMessage = strings.Join(lineErrors, "\n")    // Join all with newline

	return strings.Replace(cleanMessage, "  ", "", -1) // Remove any unnecessary whitespace
}
