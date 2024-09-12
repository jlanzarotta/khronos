package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

var BuildVersion string
var BuildDateTime string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version information",
	Long:  "Show the version information.",
	Run: func(cmd *cobra.Command, args []string) {
		runVersion(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) {
	log.Printf("  Version: " + BuildVersion + "\n" +
		"Copyright: (c) 2018-" + time.Now().Format("2006") +
		" Jeff Lanzarotta, All rights reserved\n  Born on: " + BuildDateTime + "\n")
}
