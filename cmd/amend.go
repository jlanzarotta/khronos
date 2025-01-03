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

	//FIXME	"database/sql"
	"fmt"
	"khronos/constants"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/dromara/carbon/v2"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"khronos/internal/database"
	"khronos/internal/models"
)

// amendCmd represents the amend command
var amendCmd = &cobra.Command{
	Use:   "amend",
	Args:  cobra.MaximumNArgs(1),
	Short: "Amend an entry",
	Long: `Amend is a convenient way to modify an entry, default is the last
entry.  It lets you modify the project, task, and/or datetime.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAmend(cmd, args)
	},
}

func init() {
	amendCmd.Flags().BoolP("today", constants.EMPTY, false, "List all the entries for today.")
	rootCmd.AddCommand(amendCmd)
}

func runAmend(cmd *cobra.Command, _ []string) {
	var entry models.Entry

	today, _ := cmd.Flags().GetBool("today")
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	if today {
		var input_value string = constants.EMPTY

		for {
			var t table.Writer = table.NewWriter()
			t.SetAutoIndex(true)
			t.AppendHeader(table.Row{"Project", "Task(s)", "Date/Time"})
			var entries []models.Entry = db.GetEntriesForToday(carbon.Now().StartOfDay(), carbon.Now().EndOfDay())
			for _, entry := range entries {
				t.AppendRow(table.Row{entry.Project, entry.GetTasksAsString(), entry.EntryDatetime})
			}

			log.Println(t.Render())

			fmt.Print("Please enter index number of the entry you would like to amend; otherwise, ENTER to quit...\n")
			n, _ := fmt.Scanln(&input_value)

			// If nothing was entered, break out of the loop.
			if n <= 0 {
				log.Printf("No entry amended.\n")
				return
			}

			// Validate what the user entered is actually a number.
			i, err := strconv.Atoi(input_value)
			if err != nil {
				fmt.Printf("\nPlease enter a valid value.\n\n")
				continue
			}

			// Validate that the entry was between 1 and the length of the entries.
			if i <= 0 || i > len(entries) {
				fmt.Printf("\nPlease enter a valid value.\n\n")
				continue
			}

			// Get the entry the user wants to amend.
			entry = entries[i-1]
			break
		}
	} else {
		// Get the last Entry from the database.
		entry = db.GetLastEntry()
	}

	log.Printf("%s", "Amending...\n"+entry.Dump(true, constants.INDENT_AMOUNT)+"\n\n")

	// Prompt to change project.
	newProject := prompt(constants.PROJECT_NORMAL_CASE, entry.Project)

	// If we are modifying a break, there is no need to ask for a task since
	// breaks do not have tasks.
	var newTask string = constants.EMPTY
	if !strings.EqualFold(newProject, constants.BREAK) {
		newTask = prompt(constants.TASK_NORMAL_CASE, entry.GetTasksAsString())
	}

	newNote := prompt(constants.NOTE_NORMAL_CASE, entry.Note)

	// If there was an URL, prompt to change it.
	var newURL string = constants.EMPTY
	if len(entry.GetUrlAsString()) > 0 {
		newURL = prompt(constants.URL_NORMAL_CASE, entry.GetUrlAsString())
	}

	newEntryDatetime := prompt(constants.DATE_TIME_NORMAL_CASE, entry.EntryDatetime)

	// Validate that the user entered a correctly formatted date/time.
	e := carbon.Parse(newEntryDatetime)
	if e.Error != nil {
		log.Fatalf("%s: Invalid ISO8601 date/time format.  Please try to amend again with a valid ISO8601 formatted date/time.", color.RedString(constants.FATAL_NORMAL_CASE))
	} else {
		newEntryDatetime = carbon.Parse(newEntryDatetime).ToIso8601String()
	}

	log.Printf("\n")

	// Create a table to show the old verses new values.
	var t table.Writer = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.AppendHeader(table.Row{"", "Old", "New"})
	t.AppendRow(table.Row{constants.PROJECT_NORMAL_CASE, entry.Project, newProject})
	t.AppendRow(table.Row{constants.TASK_NORMAL_CASE, entry.GetTasksAsString(), newTask})
	t.AppendRow(table.Row{constants.NOTE_NORMAL_CASE, entry.Note, newNote})

	if len(newURL) > 0 {
		t.AppendRow(table.Row{constants.URL_NORMAL_CASE, entry.GetUrlAsString(), newURL})
	}

	t.AppendRow(table.Row{constants.DATE_TIME_NORMAL_CASE, entry.EntryDatetime, newEntryDatetime})

	// Render the table.
	log.Println(t.Render())

	// Ask the user if they want to commit these changes or not.
	yesNo := yesNoPrompt("\nCommit these changes?")
	if yesNo {
		var e models.Entry
		e.Uid = entry.Uid
		e.Project = newProject
		e.Note = newNote
		e.EntryDatetime = newEntryDatetime
		e.AddEntryProperty(constants.TASK, newTask)

		if len(newURL) > 0 {
			e.AddEntryProperty(constants.URL, newURL)
		}

		db.UpdateEntry(e)

		log.Printf("%s\n", color.GreenString("Last entry amended."))
	} else {
		log.Printf("%s\n", color.YellowString("Last entry not amended."))
	}
}

func prompt(label string, value string) string {
	r := bufio.NewReader(os.Stdin)
	var s string

	fmt.Fprintf(os.Stderr, "Enter %s (empty for no change) ["+value+"] > ", label)
	s, _ = r.ReadString('\n')
	s = strings.TrimSpace(s)

	// If the result is empty, use the original passed in value.
	if len(s) <= 0 {
		s = value
	}

	return s
}
