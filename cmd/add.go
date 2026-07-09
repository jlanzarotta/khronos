/*
Copyright © 2018-2026 Jeff Lanzarotta
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
		log.Fatalf("%s: Favorite must be >= 1.\n", color.RedString(constants.FATAL_NORMAL_CASE))
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
		// index is 0-based internally; report it 1-based to match what the user
		// typed on the --favorite flag.
		log.Fatalf("%s: Favorite[%d] not found in configuration file[%s].\n", color.RedString(constants.FATAL_NORMAL_CASE), index+1, viper.ConfigFileUsed())
		os.Exit(1)
	}

	return config.Favorites[index]
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
	var addTime carbon.Carbon = *carbon.Now()

	// Get the --at flag.
	atTimeStr, _ := cmd.Flags().GetString(constants.AT)

	// Make sure there are records in the database.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var lastEntry models.Entry = db.GetLastEntry()
	if lastEntry.Uid == constants.UNKNOWN_UID {
		log.Fatalf("%s: There are no records in your database yet. To start time tracking, please perform a %s first.\n",
			color.RedString(constants.FATAL_NORMAL_CASE), color.YellowString(constants.COMMAND_HELLO))
		os.Exit(1)
	}

	// Check if we are allowed to add. Each day we must first do a HELLO to start our day.
	var allowed bool = db.AllowedToAdd()
	if allowed == false {
		log.Fatalf("%s: Unable to add entries. A `hello` must be done first thing each day.\n", color.RedString(constants.FATAL_NORMAL_CASE))
		os.Exit(1)
	}

	// Check it the --at flag was enter or not.
	if !stringUtils.IsEmpty(atTimeStr) {
		atTime, err := anytime.Parse(atTimeStr, time.Now())
		if err != nil {
			log.Fatalf("%s: Failed parsing 'at' time. %s.  For natural date examples see https://github.com/ijt/go-anytime\n",
				color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		addTime = *carbon.CreateFromStdTime(atTime)
	}

	var projectTask string = constants.EMPTY
	var description string = constants.EMPTY
	var ticket string = constants.EMPTY
	var requiredNote bool = false

	favorite, _ := cmd.Flags().GetInt(constants.FAVORITE)

	if favorite != -999 {
		// The --favorite flag is 1-based for the user (matching the displayed
		// "#" column); getFavorite indexes 0-based, so convert here.
		var fav Favorite = getFavorite(favorite - 1)
		projectTask = fav.Favorite
        description = fav.Description
		ticket = fav.Ticket
		requiredNote = fav.RequireNote
	} else {
		if len(args) > 0 {
			projectTask = args[0]
		} else {
			// Since no parameters were specified, do an interactive add using
			// the bubbles/table selector. The selector displays the favorites
			// and returns the chosen index directly, replacing the old
			// show-then-prompt-for-a-number loop.
			favs := loadFavorites()
			if len(favs) <= 0 {
				log.Fatalf("%s: No favorites found in configuration file[%s].  Unable to perform an interactive add.\n",
					color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed())
				os.Exit(1)
			}

			idx, ok, err := selectFavorite("Select a favorite to add", viper.ConfigFileUsed(), favs)
			if err != nil {
				log.Fatalf("%s: Error running favorites selector. %s\n",
					color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
				os.Exit(1)
			}

			// User cancelled (q/esc/ctrl+c or no selection): nothing to add.
			if !ok {
				log.Printf("%s\n", color.YellowString("Nothing added."))
				os.Exit(0)
			}

			var fav Favorite = getFavorite(idx)
			projectTask = fav.Favorite
            description = fav.Description
			ticket = fav.Ticket
			requiredNote = fav.RequireNote
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
			note = promptForNote(projectTask, description, true)

			// If the note is still empty, this is an indicator that the user wants to exit.
			if len(note) <= 0 {
				log.Printf("%s\n", color.YellowString("Required note not entered. Nothing added."))
				os.Exit(0)
			}
		}
	}

	// Create a new Entry.
	var entry models.Entry = models.NewEntry(constants.UNKNOWN_UID, pieces[0], note,
		addTime.ToIso8601String(carbon.UTC))

	// Populate the newly created Entry with its tasks.
	for i := 1; i < len(pieces); i += 1 {
		entry.AddEntryProperty(constants.TASK, pieces[i])
	}

	// If a Ticket was configured for this project+task, add it to the entry.
	if !stringUtils.IsBlank(ticket) {
		entry.AddEntryProperty(constants.TICKET, ticket)
		entry.AddEntryProperty(constants.PUSHED, constants.EMPTY)
	}

	// Prompt the user to make sure they really want to add this new entry.
	log.Printf("You are about to add this entry\n%s...\n\n", entry.Dump(true, constants.INDENT_AMOUNT))
	yesNo := yesNoPrompt("Continue?")
	if yesNo {
		// Yes, they want the entry added. Write the new Entry to the database.
		db.InsertNewEntry(entry)
		log.Printf("%s.\n", color.GreenString("Entry added"))
	} else {
		// No, they do not want the entry added.
		log.Printf("%s\n", color.YellowString("Nothing added."))
	}
}

func promptForNote(projectTask string, description string, required bool) string {
	var s string
	var prompt string

	if required {
		pieces := strings.Split(projectTask, "+")
		prompt = color.YellowString("Project")
		prompt += "["
		prompt += pieces[0]
		prompt += "] "
		prompt += color.YellowString("Task")
		prompt += "["
		prompt += pieces[1]

        if !stringUtils.IsEmpty(description) {
		    prompt += "] "
		    prompt += color.YellowString("Description")
		    prompt += "["
		    prompt += description
        }

		prompt += "] requires a note...\n"
	} else {
		prompt = "A note is required...\n"
	}

	prompt += "Enter note or leave blank to quit. > "

	fmt.Print(prompt)
	s, _ = readLine(stdinReader)
	s = strings.TrimSpace(s)

	return s
}

// readLine reads a single line of input from r, terminated by '\n', '\r',
// or '\r\n'. Unlike bufio.Reader.ReadString('\n'), this tolerates a bare
// '\r' - which is what a raw-mode terminal (e.g. one left in that state by
// a Bubble Tea program that hasn't fully restored cooked mode yet) sends
// for the Enter key instead of '\n'. Without this, ReadString('\n') can
// block forever waiting for a byte that never arrives.
func readLine(r *bufio.Reader) (string, error) {
	var sb strings.Builder

	for {
		b, err := r.ReadByte()
		if err != nil {
			// Return whatever we've accumulated so far (e.g. EOF mid-line).
			return sb.String(), err
		}

		if b == '\n' {
			return sb.String(), nil
		}

		if b == '\r' {
			// Peek ahead in case this is a "\r\n" pair; if so, consume the
			// '\n' too so it doesn't leak into the next read.
			next, err := r.Peek(1)
			if err == nil && len(next) == 1 && next[0] == '\n' {
				_, _ = r.ReadByte()
			}
			return sb.String(), nil
		}

		sb.WriteByte(b)
	}
}
