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
			log.Fatalf("%s: Failed parsing 'at' time. %s.  For natural date examples see https://github.com/ijt/go-anytime\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		helloTime = carbon.CreateFromStdTime(atTime)
	}

	// Create a new Entry.
	var entry models.Entry = models.NewEntry(constants.UNKNOWN_UID, constants.HELLO, constants.EMPTY, helloTime.ToRfc3339String())
	log.Printf("%s", greetings(helloTime)+" Time tracking has now started.\n")

	if viper.GetBool("debug") {
		log.Printf("helloTime=[%v] entry=[%v]\n", helloTime, entry)
	}

	// Write the new Entry to the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	db.InsertNewEntry(entry)
}
