/*
Copyright © 2023 Jeff Lanzarotta
*/
package cmd

import (
	"log"
	"os"
	"os/user"
	"time"
	"khronos/constants"
	"khronos/internal/database"
	"khronos/internal/models"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/fatih/color"
	"github.com/golang-module/carbon/v2"
	"github.com/ijt/go-anytime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Global at string for ALL commands.
var at string

// helloCmd represents the hello command
var helloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Start time tracking for the day",
	Long: `In order to have khronos start tracking time is to run this
command.  It informs khronos that you would like it to start tracking
your time.`,
	Run: func(cmd *cobra.Command, args []string) {
		runHello(cmd, args)
	},
}

func init() {
	helloCmd.Flags().StringVarP(&at, constants.AT, constants.EMPTY, constants.EMPTY, "Natural Language Time, e.g., '18 minutes ago'")
	rootCmd.AddCommand(helloCmd)
}

func greetings(c carbon.Carbon) string {
	var value string
	if c.Hour() >= 0 && c.Hour() < 12 {
		value = "Good morning, "
	} else if c.Hour() >= 12 && c.Hour() < 16 {
		value = "Good afternoon, "
	} else if c.Hour() >= 16 && c.Hour() < 21 {
		value = "Good evening, "
	} else if c.Hour() >= 21 && c.Hour() < 24 {
		value = "You are up late, "
	}

	// Get the current system user.
	currentUser, err := user.Current()
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	value = value + currentUser.Name + "."

	return value
}

func runHello(cmd *cobra.Command, _ []string) {
	// Get the current date/time.
	var helloTime carbon.Carbon = carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString(constants.AT)

	// Check it the --at flag was enter or not.
	if !stringUtils.IsEmpty(atTimeStr) {
		atTime, err := anytime.Parse(atTimeStr, time.Now())
		if err != nil {
			log.Fatalf("%s: Error parsing 'at' time. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		helloTime = carbon.CreateFromStdTime(atTime)
	}

	// Create a new Entry.
	var entry models.Entry = models.NewEntry(constants.UNKNOWN_UID, constants.HELLO, constants.EMPTY, helloTime.ToRfc3339String())
	log.Printf(greetings(helloTime) + " Time tracking starts now.\n")

	if viper.GetBool("debug") {
		log.Printf("helloTime=[%v] entry=[%v]\n", helloTime, entry)
	}

	// Write the new Entry to the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	db.InsertNewEntry(entry)
}
