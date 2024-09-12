package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configureCmd = &cobra.Command{
	Use:     "configure",
	Aliases: []string{"c", "config", "conf"},
	Short:   "Write out a YAML config file",
	Long:    "Write out a YAML config file. Print path to config file.",
	Run: func(cmd *cobra.Command, args []string) {
		runConfigure(args)
	},
}

func init() {
	rootCmd.AddCommand(configureCmd)
}

func runConfigure(args []string) {
	if len(args) != 0 {
		log.Printf("usage: khronos configure")
		os.Exit(1)
	}

	log.Printf("Config file is at \"%s\"\n", viper.GetViper().ConfigFileUsed())
	// TODO: does configuration command do anything else?

	os.Exit(0)
}
