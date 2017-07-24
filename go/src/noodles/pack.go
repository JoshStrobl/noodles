package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use:   "pack",
	Short: "Package configured assets for all or a specified project",
	Long:  "Package configured assets for all or a specified project into a distributable tarball",
	Run:   pack,
}

// pack will package configured assets for a specified project into a tarball
func pack(cmd *cobra.Command, args []string) {
	if project == "" {
		fmt.Println("Packing all the things!")
	} else {
		fmt.Printf("Packing %s\n", project)
	}
}
