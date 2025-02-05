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
	"khronos/constants"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/dromara/carbon/v2"
	"github.com/fatih/color"
	"github.com/ijt/go-anytime"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"khronos/internal/database"
	"khronos/internal/models"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [project+task...]",
	Args:  cobra.MaximumNArgs(1),
	Short: constants.ADD_SHORT_DESCRIPTION,
	Long:  constants.ADD_LONG_DESCRIPTION,
	Run: func(cmd *cobra.Command, args []string) {
		runAdd(cmd, args)
	},
}

var favorite int

func getFavorite(index int) Favorite {
	if index < 0 {
		log.Fatalf("%s: Favorite must be >= 0.\n", color.RedString(constants.FATAL_NORMAL_CASE))
		os.Exit(1)
	}

	data, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		log.Fatalf("%s: Error reading configuration file[%s]. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	var config Configuration

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("%s: Error unmarshalling configuration file[%s]. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	if index >= len(config.Favorites) {
		log.Fatalf("%s: Favorite[%d] not found in configuration file[%s].\n", color.RedString(constants.FATAL_NORMAL_CASE), index, viper.ConfigFileUsed())
		os.Exit(1)
	}

	return config.Favorites[index]
}

func getNumberOfFavorites() int {
	data, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		log.Fatalf("%s: Error reading configuration file[%s]. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	var config Configuration

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("%s: Error unmarshalling configuration file[%s]. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	return len(config.Favorites)
}

func init() {
	// Here you will define your flags and configuration settings.
	addCmd.Flags().StringVarP(&at, constants.AT, constants.EMPTY, constants.EMPTY, constants.NATURAL_LANGUAGE_DESCRIPTION)
	addCmd.Flags().StringVarP(&note, constants.NOTE, constants.EMPTY, constants.EMPTY, constants.NOTE_DESCRIPTION)
	addCmd.Flags().IntVarP(&favorite, constants.FAVORITE, constants.EMPTY, -999, "Use the specified Favorite")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) {
	// Get the current date/time.
	var addTime carbon.Carbon = carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString(constants.AT)

	// Check it the --at flag was enter or not.
	if !stringUtils.IsEmpty(atTimeStr) {
		atTime, err := anytime.Parse(atTimeStr, time.Now())
		if err != nil {
			log.Fatalf("%s: Failed parsing 'at' time. %s.  For natural date examples see https://github.com/ijt/go-anytime\n",
				color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		addTime = carbon.CreateFromStdTime(atTime)
	}

	var projectTask string = constants.EMPTY
	var url string = constants.EMPTY
	var requiredNote bool = false

	favorite, _ := cmd.Flags().GetInt(constants.FAVORITE)

	if favorite != -999 {
		var fav Favorite = getFavorite(favorite)
		projectTask = fav.Favorite
		url = fav.URL
	} else {
		if len(args) > 0 {
			projectTask = args[0]
		} else {
			for {
				if getNumberOfFavorites() <= 0 {
					log.Fatalf("%s: No favorites found in configuration file[%s].  Unable to perform an interactive add.\n",
						color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed())
					os.Exit(1)
				}

				// Since no parameters were specified, do an interactive add.
				showFavorites()

				// Prompt the user for the index number of the filename they would like to send.
				r := bufio.NewReader(os.Stdin)

				fmt.Fprintf(os.Stderr, "\nPlease enter the number of the favorite to add; otherwise, [Return] to quit. > ")
				var s, _ = r.ReadString('\n')
				s = strings.TrimSpace(s)

				// If the result is empty, the user wants to quit.
				if len(s) <= 0 {
					log.Printf("%s\n", color.YellowString("Nothing added."))
					os.Exit(0)
				}

				// Convert the string to an integer, thus validating the user entered a number.
				i, err := strconv.Atoi(s)
				if err != nil {
					log.Printf("Invalid number entered.\n")
					continue
				}

				var fav Favorite = getFavorite(i)
				projectTask = fav.Favorite
				url = fav.URL
				requiredNote = fav.RequireNote
				break
			}
		}
	}

	// Split the project/task into pieces.
	var pieces []string = strings.Split(projectTask, constants.TASK_DELIMITER)
	if len(pieces) < 2 {
		log.Fatalf("%s: Unable to parsing 'project+task'.  Malformed project+task.\n", color.RedString(constants.FATAL_NORMAL_CASE))
		os.Exit(1)
	}

	// Check if the note was empty and the require_note flag is globally set or
	// set on the favorite.  If so, require the note.
	if stringUtils.IsEmpty(note) {
		var globalRequired bool = viper.GetBool(constants.REQUIRE_NOTE)
		if globalRequired || requiredNote {
			note = promptForNote(true)

			// If the note is still empty, this is an indicator that the user wants to exit.
			if len(note) <= 0 {
				log.Printf("%s\n", color.YellowString("Nothing added."))
				os.Exit(0)
			}
		}
	}

	// Create a new Entry.
	var entry models.Entry = models.NewEntry(constants.UNKNOWN_UID, pieces[0], note,
		addTime.ToRfc3339String())

	// Populate the newly created Entry with its tasks.
	for i := 1; i < len(pieces); i += 1 {
		entry.AddEntryProperty(constants.TASK, pieces[i])
	}

	// If a URL was configured for this project+task, add it to the entry.
	if len(url) > 0 {
		entry.AddEntryProperty(constants.URL, url)
	}

	log.Printf("%s%s.\n", color.GreenString(constants.ADDING), entry.Dump(true, constants.INDENT_AMOUNT))

	// Write the new Entry to the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	db.InsertNewEntry(entry)
}

func promptForNote(required bool) string {
	r := bufio.NewReader(os.Stdin)
	var s string
	var prompt string

	if required {
		prompt = "A note is required.  "
	}

	prompt += "Enter note or leave blank to quit. > "

	fmt.Print(prompt)
	s, _ = r.ReadString('\n')
	s = strings.TrimSpace(s)

	return s
}
