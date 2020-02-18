package main

// Vendoring Functionality

import (
	"github.com/JoshStrobl/trunk"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
)

var vendorCmd = &cobra.Command{
	Use:   "vendor",
	Short: "Vendor external dependencies for your project",
	Long:  "Vendor external dependencies for your project",
	Run:   vendor,
}

// vendor command handler
func vendor(cmd *cobra.Command, args []string) {

}

// add is responsible for adding a vendor based on its name and inputted values
func add(name string) {
	if name != "" {
		coreutils.InputMessage("Please enter the repository URL and press [Enter]:")
	} else {
		trunk.LogErr("Please provide a name for this vendor, such as the repository name.")
	}
}

// fetch is responsible for fetching / checking out a vendor based on its name
func fetch(name string) {
	if name != "" {

	} else {
		trunk.LogErr("Please provide a vendor to fetch.")
	}
}

// delete is responsible for deleting a vendor based on its name
func delete(name string) {
	if name != "" {

	} else {
		trunk.LogErr("Please provide a vendor to delete.")
	}
}
