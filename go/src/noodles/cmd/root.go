package cmd

import (
	"github.com/spf13/cobra"
)

// Global Variables
var project string

var RootCmd = &cobra.Command{
	Use: "noodles",
	Short: "noodles is an opinionated manager for web apps.",
	Long: `noodles is an opinionated manager for web applications, enabling various functionality such as:
	- basic dependency management for built-in plugin support
	- compilation of project(s) in a configurable, ordered manner
	- configurable packing of project assets for distribution`,
}

// Initialize RootCmd bits
func init() {
	RootCmd.PersistentFlags().StringVarP(&project, "project", "p", "", "Project to apply for specific commands")
}