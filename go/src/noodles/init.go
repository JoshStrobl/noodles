package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stroblindustries/coreutils"
	"os"
)

func initNoodles(cmd *cobra.Command, args []string) {
	if configInfo, statErr := os.Stat("noodles.yml"); os.IsExist(statErr) { // Check if noodles.yml already exists
		if configInfo.Size() != 0 { // If the size of the file is greater than 0, meaning there is potentially content
			fmt.Println("noodles.yml already exists and appears to have content. Exiting.")
			return
		}
	}

	noodles.name = coreutils.InputMessage("Name of Noodles Project")
	fmt.Printf("%v", noodles)
}