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

	//FIXME	"database/sql"
	"fmt"
	"khronos/constants"
	"log"
	"os"
	"strings"

	"github.com/agrison/go-commons-lang/stringUtils"
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
	Short: constants.AMEND_SHORT_DESCRIPTION,
	Long:  constants.AMEND_LONG_DESCRIPTION,
	Run: func(cmd *cobra.Command, args []string) {
		runAmend(cmd, args)
	},
}

func init() {
	amendCmd.Flags().BoolP("today", constants.EMPTY, false, "List all the entries for today.")
	amendCmd.Flags().StringVarP(&givenDate, constants.FLAG_DATE, constants.EMPTY, constants.EMPTY, "List all the entries for the given day in "+constants.DATE_FORMAT_YYYY_MM_DD+" format.")
	amendCmd.MarkFlagsMutuallyExclusive("today", constants.FLAG_DATE)
	rootCmd.AddCommand(amendCmd)
}

func runAmend(cmd *cobra.Command, _ []string) {
	var entry models.Entry

	today, _ := cmd.Flags().GetBool("today")
	givenDate, _ := cmd.Flags().GetString(constants.FLAG_DATE)

	db := database.New(viper.GetString(constants.DATABASE_FILE))
	if today || !stringUtils.IsEmpty(givenDate) {
		var entries []models.Entry

		if today {
			entries = db.GetEntriesForToday(*carbon.Now().StartOfDay(), *carbon.Now().EndOfDay())
		} else {
			entries = db.GetEntriesForToday(*carbon.Parse(givenDate).StartOfDay(), *carbon.Parse(givenDate).EndOfDay())
		}

		if len(entries) == 0 {
			log.Printf("%s\n", color.YellowString("No entries found."))
			return
		}

		// Show the entries in an interactive selector and let the user pick one.
		idx, ok, err := selectEntry("Select an entry to amend", entries)
		if err != nil {
			log.Fatalf("%s: Error running entry selector. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		if !ok {
			// User cancelled - nothing to amend.
			log.Printf("%s\n", color.YellowString("No entry amended."))
			return
		}

		entry = entries[idx]
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

	// If there was an TICKET, prompt to change it.
	var newTicket string = constants.EMPTY
	if len(entry.GetTicketAsString()) > 0 {
		newTicket = prompt(constants.TICKET_NORMAL_CASE, entry.GetTicketAsString())
	}

	newEntryDatetime := prompt(constants.DATE_TIME_NORMAL_CASE, carbon.Parse(entry.EntryDatetime).ToIso8601String(carbon.Local))

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

	if len(newTicket) > 0 {
		t.AppendRow(table.Row{constants.TICKET_NORMAL_CASE, entry.GetTicketAsString(), newTicket})
	}

	t.AppendRow(table.Row{constants.DATE_TIME_NORMAL_CASE,
		carbon.Parse(entry.EntryDatetime).ToIso8601String(carbon.Local),
		carbon.Parse(newEntryDatetime).ToIso8601String(carbon.Local)})

	// Render the table.
	log.Println(t.Render())

	// Ask the user if they want to commit these changes or not.
	yesNo := yesNoPrompt("\nCommit these changes?")
	if yesNo {
		var e models.Entry
		e.Uid = entry.Uid
		e.Project = newProject
		e.Note = newNote
		e.EntryDatetime = carbon.Parse(newEntryDatetime).ToIso8601String()
		e.AddEntryProperty(constants.TASK, newTask)

		if len(newTicket) > 0 {
			e.AddEntryProperty(constants.TICKET, newTicket)
		}

		db.UpdateEntry(e)

		log.Printf("%s\n", color.GreenString("Entry amended."))
	} else {
		log.Printf("%s\n", color.YellowString("Entry NOT amended."))
	}
}

func prompt(label string, value string) string {
	var s string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Fprintf(os.Stderr, "Enter %s (empty for no change) ["+value+"] > ", label)
	if !scanner.Scan() {
		s = scanner.Text()

		// If the result is empty, use the original passed in value.
		if s == constants.EMPTY {
			s = value
		}
	} else {
		s = value
	}
	return s
}
