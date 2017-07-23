package cmd

import (
	//"fmt"
	"github.com/spf13/cobra"
	//"github.com/stroblindustries/coreutils"
	//"gopkg.in/yaml.v2"
)

var initCmd = &cobra.Command{
	Use: "init",
	Short: "initialize noodles",
	Long: "initialize noodles by generating a basic YAML configuration file",
	Run: initNoodles,
}

func initNoodles(cmd *cobra.Command, args []string) {

}