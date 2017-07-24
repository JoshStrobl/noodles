package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
	"strconv"
)

func initNoodles(cmd *cobra.Command, args []string) {
	if configInfo, statErr := os.Stat("noodles.yml"); statErr == nil { // Check if noodles.yml already exists
		if configInfo.Size() != 0 { // If the size of the file is greater than 0, meaning there is potentially content
			fmt.Println("noodles.yml already exists and appears to have content. Exiting.")
			return
		}
	}

	noodles.Name = coreutils.InputMessage("Name of Noodles Project")
	noodles.Description = coreutils.InputMessage("Description of this project")
	noodles.License = coreutils.InputMessage("License")

	for noodles.Version == 0 {
		version := coreutils.InputMessage("Version")

		if num, convertErr := strconv.ParseFloat(version, 64); convertErr == nil { // Convert the version to a float64, if it's valid
			noodles.Version = num
		} else {
			fmt.Println("Invalid Version Number. Please try again.")
		}
	}

	if saveErr := noodles.Save(); saveErr == nil { // Save the config
		fmt.Println("Noodles is now inited.")
	} else { // Failed to save
		fmt.Println(saveErr.Error())
	}
}