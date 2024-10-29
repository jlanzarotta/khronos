package cmd

import (
	"log"
	"khronos/constants"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

// webCmd represents the web command.
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Open the Khronos website in your default browser",
	Long:  "Open the Khronos website in your default browser.",
	Run: func(cmd *cobra.Command, args []string) {
		runWeb(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(webCmd)
}

func runWeb(_ *cobra.Command, _ []string) {
	log.Printf("Opening the %s website in your default browser...\n", constants.APPLICATION_NAME)
	var error = browser.OpenURL(constants.WEB_SITE)
	if error != nil {
		log.Printf("Unable to open URL[%s].  Error: %s\n", constants.WEB_SITE, error)
	}
}
