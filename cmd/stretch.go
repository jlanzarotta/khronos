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
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"khronos/constants"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/fatih/color"
	"github.com/dromara/carbon/v2"
	"github.com/ijt/go-anytime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"khronos/internal/database"
	"khronos/internal/models"
)

// stretchCmd represents the stretch command
var stretchCmd = &cobra.Command{
	Use:   "stretch last project",
	Short: constants.STRETCH_SHORT_DESCRIPTION,
	Long:  constants.STRETCH_LONG_DESCRIPTION,
	Run: func(cmd *cobra.Command, args []string) {
		runStretch(cmd, args)
	},
}

func init() {
	stretchCmd.Flags().StringVarP(&at, constants.AT, constants.EMPTY, constants.EMPTY, "Natural Language Time, e.g., '18 minutes ago'")
	rootCmd.AddCommand(stretchCmd)
}

func runStretch(cmd *cobra.Command, _ []string) {
	// Get the current date/time.
	var stretchTime carbon.Carbon = carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString(constants.AT)

	// Check it the --at flag was enter or not.
	if !stringUtils.IsEmpty(atTimeStr) {
		atTime, err := anytime.Parse(atTimeStr, time.Now())
		if err != nil {
			log.Fatalf("%s: Failed parsing 'at' time. %s\n", color.RedString("Fatal"), err.Error())
			os.Exit(1)
		}

		stretchTime = carbon.CreateFromStdTime(atTime)
	}

	// Get the last Entry from the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var entry models.Entry = db.GetLastEntry()

	// Create the prompt.
	var prompt string = "Would you like to stretch\n" + entry.Dump(true, constants.INDENT_AMOUNT)
	prompt = prompt + "\n\nto " + stretchTime.ToCookieString() + "?"

	// Ask the user if they actually want to stretch the last entry or not.
	yesNo := yesNoPrompt(prompt)
	if yesNo {
		// Yes was enter, so update the latest.
		var e models.Entry
		e.Uid = entry.Uid
		e.EntryDatetime = stretchTime.ToIso8601String()
		db.UpdateEntry(e)

		log.Printf("%s\n", color.GreenString("Last entry was stretched."))
	} else {
		log.Printf("%s\n", color.YellowString("Last entry was NOT stretched."))
	}
}

func yesNoPrompt(label string) bool {
	choices := "Y/N (yes/no)"

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) > ", label, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		} else if s == "n" || s == "no" {
			return false
		} else {
			return false
		}
	}
}
