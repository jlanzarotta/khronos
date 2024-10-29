package cmd

import (
	"log"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// editCmd represents the edit command.
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open the Khronos configuration file in your default editor",
	Long:  "Open the Khronos configuration file in your default editor.",
	Run: func(cmd *cobra.Command, args []string) {
		runEdit(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}

func runEdit(_ *cobra.Command, _ []string) {
	log.Printf("Opening the %s file in your default editor...\n", viper.ConfigFileUsed())
	exePath := "c:\\windows\\system32\\notepad.exe"
	cmd := exec.Command(exePath, viper.ConfigFileUsed())
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}
