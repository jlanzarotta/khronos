/*
Copyright © 2023 Jeff Lanzarotta
*/
package cmd

import (
	"khronos/constants"
	"khronos/internal/database"
	"khronos/internal/models"
	"log"
	"os"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/fatih/color"
	"github.com/dromara/carbon/v2"
	"github.com/ijt/go-anytime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// breakCmd represents the break command
var breakCmd = &cobra.Command{
	Use:   "break",
	Short: "Add a break",
	Long: `If you just spent time on break, use this command to add that time
to the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		runBreak(cmd, args)
	},
}

func init() {
	breakCmd.Flags().StringVarP(&at, constants.AT, constants.EMPTY, constants.EMPTY, constants.NATURAL_LANGUAGE_DESCRIPTION)
	breakCmd.Flags().StringVarP(&note, constants.NOTE, constants.EMPTY, constants.EMPTY, constants.NOTE_DESCRIPTION)
	rootCmd.AddCommand(breakCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// breakCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// breakCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runBreak(cmd *cobra.Command, _ []string) {
	// Get the current date/time.
	var breakTime carbon.Carbon = carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString(constants.AT)

	// Check it the --at flag was enter or not.
	if !stringUtils.IsEmpty(atTimeStr) {
		atTime, err := anytime.Parse(atTimeStr, time.Now())
		if err != nil {
			log.Fatalf("%s: Failed parsing 'at' time. %s.  For natural date examples see https://github.com/ijt/go-anytime\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		breakTime = carbon.CreateFromStdTime(atTime)
	}

	// Create a new Entry.
	var entry models.Entry = models.NewEntry(constants.UNKNOWN_UID, constants.BREAK, note,
		breakTime.ToRfc3339String())

	log.Printf("%s %s.\n", color.GreenString(constants.ADDING), entry.Dump(false, 0))

	// Write the new Entry to the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	db.InsertNewEntry(entry)
}
