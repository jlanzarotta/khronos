/*
Copyright © 2018-2025 Jeff Lanzarotta
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"khronos/constants"
	"khronos/internal/database"
	"khronos/internal/models"
	"log"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/dromara/carbon/v2"
	"github.com/fatih/color"
	"github.com/ijt/go-anytime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Global at string for ALL commands.
var at string

// helloCmd represents the hello command
var helloCmd = &cobra.Command{
	Use:   constants.COMMAND_HELLO,
	Short: constants.HELLO_SHORT_DESCRIPTION,
	Long:  constants.HELLO_LONG_DESCRIPTION,
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

	// Convert to local time so we get the correct hour.
	var localTime = c.SetTimezone(carbon.Local)

	if c.Hour() >= 0 && localTime.Hour() < 12 {
		value = "Good morning, "
	} else if localTime.Hour() >= 12 && localTime.Hour() < 16 {
		value = "Good afternoon, "
	} else if localTime.Hour() >= 16 && localTime.Hour() < 21 {
		value = "Good evening, "
	} else if localTime.Hour() >= 21 && localTime.Hour() < 24 {
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
	var helloTime carbon.Carbon = *carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString(constants.AT)

	// Check it the --at flag was enter or not.
	if !stringUtils.IsEmpty(atTimeStr) {
		atTime, err := anytime.Parse(atTimeStr, time.Now())
		if err != nil {
			log.Fatalf("%s: Failed parsing 'at' time. %s.  For natural date examples see https://github.com/ijt/go-anytime\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		helloTime = *carbon.CreateFromStdTime(atTime)
	}

	// Create a new Entry.
	var entry models.Entry = models.NewEntry(constants.UNKNOWN_UID, constants.HELLO, constants.EMPTY,
		helloTime.ToIso8601String(carbon.UTC))

	// Get the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))

	// Get the last Entry.
	var lastEntry models.Entry = db.GetLastEntry()
	if lastEntry.Uid != constants.UNKNOWN_UID {
		// Check if the last entry was a HELLO. If is was, was it on the day as this
		// new entry? If so, reject the new attempt to add a HELLO.
		if strings.EqualFold(lastEntry.Project, constants.HELLO) {
			var lastDateTime carbon.Carbon = *carbon.Parse(lastEntry.EntryDatetime)
			var helloDateTime carbon.Carbon = *carbon.Parse(entry.EntryDatetime)
			var diff int64 = lastDateTime.DiffAbsInDays(&helloDateTime)

			if diff == 0 {
				log.Printf("%s\n", color.YellowString("No need to start time tracking for today, as it was already started at "+lastDateTime.String()+"."))
				return
			}
		}
	}

	log.Printf("%s", greetings(helloTime)+" Time tracking has now started.\n")

	if viper.GetBool(constants.DEBUG) {
		log.Printf("helloTime=[%v] entry=[%v]\n", helloTime, entry)
	}

	// Write the new Entry to the database.
	db.InsertNewEntry(entry)
}
