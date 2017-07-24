package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"regexp"
	"strings"
)

// lint will validate noodles.yml
func lint(cmd *cobra.Command, args []string) {
	readErr := noodles.Read() // Read the config

	if readErr == nil {
		fmt.Println("noodles.yml appears correct.")
	} else {
		cleanMessage := CleanLintErrors(readErr.Error())
		fmt.Printf("noodles.yml appears to have the following issue(s):\n%s\n", cleanMessage)
	}
}

// CleanLintErrors will remove some verbosity from any unmarshalling error
func CleanLintErrors(err string) string {
	cleanMessage := strings.Replace(err, "unmarshal", "convert", -1) // Change "unmarshal" to a human language
	cleanMessage = strings.Replace(cleanMessage, "!!", "", -1)
	cleanMessage = strings.Replace(cleanMessage, "`", "", -1) // Remove any ` wrapping types

	re := regexp.MustCompile(`line\s\d+:\s[\s\S]+$`) // Only get line N: message
	lineErrors := re.FindAllString(cleanMessage, -1) // Find all strings
	cleanMessage = strings.Join(lineErrors, "\n") // Join all with newline

	return strings.Replace(cleanMessage, "  ", "", -1) // Remove any unnecessary whitespace
}