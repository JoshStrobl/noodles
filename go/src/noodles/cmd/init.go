package cmd

import (
	//"fmt"
	"github.com/spf13/cobra"
	//"github.com/stroblindustries/coreutils"
	//"gopkg.in/yaml.v2"
)

var initCmd = &cobra.Command{
	Use: "init",
	Short: "Initialize noodles",
	Long: "Initialize noodles by generating a basic YAML configuration file",
	Run: initNoodles,
}

func init() {
	RootCmd.AddCommand(initCmd)
}

func initNoodles(cmd *cobra.Command, args []string) {

}