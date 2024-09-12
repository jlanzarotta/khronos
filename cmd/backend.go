package cmd

import (
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var backendCmd = &cobra.Command{
	Use:     "backend",
	Aliases: []string{"b", "back"},
	Args:    cobra.ExactArgs(0),
	Short:   "Open a sqlite shell to the database",
	Long:    "Open a sqlite shell to the database.",
	Run: func(cmd *cobra.Command, args []string) {
		runBackend(args)
	},
}

func init() {
	rootCmd.AddCommand(backendCmd)
}

func runBackend(_ []string) {
	cmd := exec.Command("sqlite3", viper.GetString("database_file"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	err := cmd.Run()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}
