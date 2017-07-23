package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
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
	noodles.Version = coreutils.InputMessage("Version")

	if saveErr := noodles.Save(); saveErr == nil { // Save the config
		fmt.Println("Noodles is now inited.")
	} else { // Failed to save
		fmt.Println(saveErr.Error())
	}
}