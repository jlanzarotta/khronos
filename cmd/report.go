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
	"bytes"
	"fmt"
	"io"
	"khronos/constants"
	"khronos/internal/database"
	"khronos/internal/jira"
	"khronos/internal/models"
	"khronos/internal/rest"
	"khronos/internal/util"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/dromara/carbon/v2"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var from string
var to string
var givenDate string
var project string
var daysOfWeek = map[string]time.Weekday{}
var roundToMinutes int64
var exportFilename string = constants.EMPTY
var _cmd *cobra.Command
var exportType = models.ExportTypeCSV
var startEndTimeFormat string = constants.CARBON_START_END_TIME_FORMAT
var pushCredentials models.Credentials
var SUB int = 40

// reportCmd represents the report command.
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: constants.REPORT_SHORT_DESCRIPTION,
	Long:  constants.REPORT_LONG_DESCRIPTION,
	Run: func(cmd *cobra.Command, args []string) {
		runReport(cmd, args)
	},
}

func separator(input string) string {
	// total '=' chars we need, accounting for the two spaces around input
	total := terminalWidth - len(input) - 2
	if total < 0 {
		total = 0
	}
	left := total / 2
	right := total - left // gets the leftover when total is odd

	return fmt.Sprintf("%s %s %s",
		strings.Repeat("=", left),
		input,
		strings.Repeat("=", right))
}

func dateRange(start *carbon.Carbon, end *carbon.Carbon) {
	dayOfWeek, err := parseWeekday(viper.GetString(constants.WEEK_START))
	if err != nil {
		log.Fatalf("%s: %s is an invalid day of week.  Please correct your configuration.\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.GetString(constants.WEEK_START))
		os.Exit(1)
	}

	*start = *start.SetWeekStartsAt(dayOfWeek).StartOfWeek().StartOfDay()
	*end = *start.Copy()
	*end = *end.AddDays(6).EndOfDay()
}

func export(title string, t table.Writer) {
	exporting, _ := _cmd.Flags().GetBool(constants.EXPORT)
	if exporting {
		typeStr, _ := _cmd.Flags().GetString(constants.EXPORT_TYPE)
		if len(strings.TrimSpace(exportFilename)) == 0 {
			// Create our new export file.
			exportFilename = constants.APPLICATION_NAME_LOWERCASE + "_report_" + carbon.Now(carbon.Local).ToShortDateTimeString()
			if typeStr == string(models.ExportTypeCSV) {
				exportFilename += ".csv"
			} else if typeStr == string(models.ExportTypeHTML) {
				exportFilename += ".html"
			} else {
				exportFilename += ".md"
			}

			_, err := os.OpenFile(exportFilename, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}

		// Open the file for appending.
		file, err := os.OpenFile(exportFilename, os.O_APPEND|os.O_WRONLY, 0)
		if err != nil {
			log.Fatal(err)
		}

		// Remember to close the file.
		defer file.Close()

		file.WriteString(title + "\n")

		// Render the table to the file.
		if typeStr == string(models.ExportTypeCSV) {
			_, err = file.WriteString(t.RenderCSV())
		} else if typeStr == string(models.ExportTypeHTML) {
			_, err = file.WriteString(t.RenderHTML())
		} else {
			_, err = file.WriteString(t.RenderMarkdown())
		}

		if err != nil {
			log.Fatal(err)
		}

		file.WriteString("\n")
	}
}

func init() {
	reportCmd.Flags().BoolP(constants.FLAG_NO_ROUNDING, constants.EMPTY, false, "Reports all durations in their unrounded form.")
	reportCmd.Flags().BoolP(constants.FLAG_CURRENT_WEEK, constants.EMPTY, false, "Report on the current week's entries.")
	reportCmd.Flags().BoolP(constants.FLAG_PREVIOUS_WEEK, constants.EMPTY, false, "Report on the previous week's entries.")
	reportCmd.Flags().BoolP(constants.FLAG_YESTERDAY, constants.EMPTY, false, "Report on yesterday's entries.")
	reportCmd.Flags().BoolP(constants.FLAG_TODAY, constants.EMPTY, false, "Report on today's entries.")
	reportCmd.Flags().BoolP(constants.FLAG_PUSH, constants.EMPTY, false, "Push the reported entries that have not yet been pushed.")
	reportCmd.Flags().StringVarP(&project, constants.FLAG_PROJECT, constants.EMPTY, constants.EMPTY, "Report on a specific project.")
	reportCmd.Flags().StringVarP(&givenDate, constants.FLAG_DATE, constants.EMPTY, constants.EMPTY, "Report on the given day's entries in "+constants.DATE_FORMAT_YYYY_MM_DD+" format.")
	reportCmd.Flags().BoolP(constants.FLAG_LAST_ENTRY, constants.EMPTY, false, "Display the last entry's information.")
	reportCmd.Flags().StringVarP(&from, constants.FLAG_FROM, constants.EMPTY, constants.EMPTY, "Specify an inclusive start date to report in "+constants.DATE_FORMAT_YYYY_MM_DD+" format.")
	reportCmd.Flags().StringVarP(&to, constants.FLAG_TO, constants.EMPTY, constants.EMPTY, "Specify an inclusive end date to report in "+constants.DATE_FORMAT_YYYY_MM_DD+" format.  If this is a day of the week, then it is the next occurrence from the start date of the report, including the start date itself.")
	reportCmd.MarkFlagsRequiredTogether(constants.FLAG_FROM, constants.FLAG_TO)
	reportCmd.Flags().BoolP(constants.EXPORT, constants.EMPTY, false, "Export to file.")
	reportCmd.Flags().Var(&exportType, constants.EXPORT_TYPE, `Type of export file.  Allowed values: "csv", "html" or "md"`)
	reportCmd.MarkFlagsRequiredTogether(constants.EXPORT, constants.EXPORT_TYPE)
	rootCmd.AddCommand(reportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Populate days of week.
	daysOfWeek["sunday"] = carbon.Sunday
	daysOfWeek["monday"] = carbon.Monday
	daysOfWeek["tuesday"] = carbon.Tuesday
	daysOfWeek["wednesday"] = carbon.Wednesday
	daysOfWeek["thursday"] = carbon.Thursday
	daysOfWeek["friday"] = carbon.Friday
	daysOfWeek["saturday"] = carbon.Saturday
}

func parseWeekday(v string) (time.Weekday, error) {
	if d, ok := daysOfWeek[strings.ToLower(v)]; ok {
		return d, nil
	}
	return -1, fmt.Errorf("invalid weekday '%s'", v)
}

// plural formats count followed by singular, pluralizing the noun for any
// count other than one. A zero count returns an empty string so that
// zero-valued components drop out of a composed duration entirely.
func plural(count int, singular string) (result string) {
	if count == 0 {
		return ""
	}

	if (count == 1) || (count == -1) {
		result = strconv.Itoa(count) + " " + singular + " "
	} else {
		result = strconv.Itoa(count) + " " + singular + "s "
	}

	return
}

func reportByDay(entries []models.Entry) {
	var display_by_day_totals bool = viper.GetBool(constants.DISPLAY_BY_DAY_TOTALS)
	log.Printf("\n")
	log.Printf("%s\n", separator(" By Day "))
	log.Printf("\n")

	// Consolidate by day.
	var consolidatedByDay map[string]map[string]models.Entry = make(map[string]map[string]models.Entry)
	for _, entry := range entries {
		var task = entry.GetTasksAsString()
		var dayKey string = carbon.Parse(entry.EntryDatetime).SetTimezone(carbon.Local).Format(constants.CARBON_DATE_FORMAT)
		consolidatedDay, found := consolidatedByDay[dayKey]
		if found {
			consolidatedProject, found := consolidatedDay[entry.Project]
			if found {
				if len(task) > 0 {
					consolidatedProject.AddEntryProperty(constants.TASK, task)
				}

				// Add the rounded durations together.
				consolidatedProject.Duration += util.Round(roundToMinutes, entry.Duration)

				// Replace the consolidated entry.
				consolidatedByDay[dayKey][entry.Project] = consolidatedProject
			} else {
				var newEntry models.Entry = models.NewEntry(entry.Uid, entry.Project, entry.Note, entry.EntryDatetime)
				newEntry.Duration = util.Round(roundToMinutes, entry.Duration)
				newEntry.Properties = entry.Properties

				// Add the new entry.
				consolidatedByDay[dayKey][entry.Project] = newEntry
			}
		} else {
			// Since the EntryDatetime was not found, add it.
			var newEntry models.Entry = models.NewEntry(entry.Uid, entry.Project, entry.Note, entry.EntryDatetime)
			newEntry.Duration = util.Round(roundToMinutes, entry.Duration)
			newEntry.Properties = entry.Properties

			// Add the new entry.
			consolidatedByDay[dayKey] = make(map[string]models.Entry)
			consolidatedByDay[dayKey][newEntry.Project] = newEntry
		}
	}

	// Since maps are not sorted in go... why, I have no idea, you need to first
	// sort the keys and then access the map via those sorted keys.
	var sortedKeys []string = make([]string, 0, len(consolidatedByDay))
	for key := range consolidatedByDay {
		sortedKeys = append(sortedKeys, key)
	}
	sort.SliceStable(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	// Create and configure the table.
	var t table.Writer = table.NewWriter()
	SetReportTableStyle(t)

	// Set a couple of the columns to fixed values and align the footer.
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Number:   1,
			WidthMin: 10,
			WidthMax: 10,
		},
		{
			Number:   4,
			WidthMin: 5,
			WidthMax: 5,
			Align:    text.AlignRight,
		},
	})

	t.AppendHeader(table.Row{constants.DATE_NORMAL_CASE, constants.PROJECT_NORMAL_CASE, constants.TASKS_NORMAL_CASE, constants.EMPTY,
		constants.DURATION_NORMAL_CASE})

	// Add each row to the table.
	for _, i := range sortedKeys {
		var day map[string]models.Entry = consolidatedByDay[i]
		var totalPerDay int64 = 0

		// The projects within a day live in a map too, so they need the same
		// sorted key treatment.  Otherwise the rows within a day come out in a
		// random order that changes from run to run.
		var sortedProjects []string = make([]string, 0, len(day))
		for project := range day {
			sortedProjects = append(sortedProjects, project)
		}
		sort.SliceStable(sortedProjects, func(x, y int) bool { return sortedProjects[x] < sortedProjects[y] })

		for _, p := range sortedProjects {
			var v models.Entry = day[p]
			t.AppendRow(table.Row{i, p, v.GetTasksAsString(), constants.EMPTY, secondsToHuman(v.Duration, true)})

			// v.Duration was already rounded during consolidation above.
			totalPerDay += v.Duration
		}

		if display_by_day_totals {
			t.AppendSeparator()
			t.AppendRow(table.Row{constants.EMPTY, constants.EMPTY, constants.EMPTY, constants.TOTAL, stringUtils.UpperCase(secondsToHuman(totalPerDay, true))})
			t.AppendSeparator()
		}
	}

	// Render the table.
	log.Println(t.Render())

	// Export table if needed.
	export("report by day", t)
}

// Column sizing for the by entry report.  Each column is measured against the
// content it actually holds, and only when the total will not fit in the
// terminal are the variable width columns shrunk to make room.
const (
	columnOverhead int = 3 // The " | " drawn between two columns.

	minimumProjectColumnWidth int = 8
	minimumTaskColumnWidth    int = 10
	minimumNoteColumnWidth    int = 12
)

// entryRow holds one rendered row of the by entry report.  The rows are built
// before the table is configured so that the column widths can be measured
// from the real content rather than guessed at with fixed values.
type entryRow struct {
	date     string
	startEnd string
	duration string
	project  string
	task     string
	pushed   string
	note     string
}

// widen returns the larger of the current width and the display width of the
// given value.
func widen(current int, value string) int {
	if width := runewidth.StringWidth(value); width > current {
		return width
	}

	return current
}

// fitColumns scales the variable width columns so that together they occupy
// exactly available characters, growing them when the terminal has room to
// spare and shrinking them when it does not.  Each column keeps at least its
// minimum and the rest of the space is handed out in proportion to how much
// each column actually wants, so the note column, which is the one that runs
// long, ends up with the lion's share.
func fitColumns(available int, widths []int, minimums []int) {
	var total int = 0
	var minimum int = 0

	// No column may drop below its minimum, so start from there.
	for i := range widths {
		if widths[i] < minimums[i] {
			widths[i] = minimums[i]
		}

		total += widths[i]
		minimum += minimums[i]
	}

	// There is not even room for the minimums, so this is as small as the
	// report can get.  The terminal is simply too narrow.
	if available <= minimum {
		copy(widths, minimums)
		return
	}

	var slack int = available - minimum
	var wanted int = total - minimum

	// Every column is already sitting at its minimum, so there is nothing to
	// scale in proportion to.  Give the slack to the last column.
	if wanted <= 0 {
		copy(widths, minimums)
		widths[len(widths)-1] += slack

		return
	}

	var used int = 0
	for i := range widths {
		widths[i] = minimums[i] + ((widths[i] - minimums[i]) * slack / wanted)
		used += widths[i]
	}

	// Integer division leaves a few characters unassigned, so give them to the
	// last column, which is the one most likely to want them.
	widths[len(widths)-1] += available - used
}

func reportByEntry(entries []models.Entry) {
	log.Printf("\n")
	log.Printf("%s\n", separator(" By Entry "))
	log.Printf("\n")

	// Build the rows up front and measure each column against its content and
	// its header.
	var rows []entryRow = make([]entryRow, 0, len(entries))
	var ticketFound bool = false

	var dateWidth int = widen(0, constants.DATE_NORMAL_CASE)
	var startEndWidth int = widen(0, constants.START_END_NORMAL_CASE)
	var durationWidth int = widen(0, constants.DURATION_NORMAL_CASE)
	var pushedWidth int = widen(0, constants.PUSHED_NORMAL_CASE)
	var projectWidth int = widen(0, constants.PROJECT_NORMAL_CASE)
	var taskWidth int = widen(0, constants.TASK_NORMAL_CASE)
	var noteWidth int = widen(0, constants.NOTE_NORMAL_CASE)

	for _, entry := range entries {
		if !stringUtils.IsEmpty(entry.GetTicketAsString()) {
			ticketFound = true
		}

		var end carbon.Carbon = *carbon.Parse(entry.EntryDatetime).SetTimezone(carbon.Local)
		var start carbon.Carbon = *carbon.Parse(entry.EntryDatetime).SetTimezone(carbon.Local).SubSeconds(int(entry.Duration))

		// Entries without a ticket cannot be pushed at all, so their pushed
		// column stays empty.  The rest report a simple yes or no.  The
		// timestamp of the push itself is available from the show command and
		// is far too wide to earn a place in this table.
		var pushed string = constants.EMPTY
		if !stringUtils.IsBlank(entry.GetTicketAsString()) {
			pushed = constants.NO_NORMAL_CASE
			if !stringUtils.IsBlank(entry.GetPushedAsString()) {
				pushed = constants.YES_NORMAL_CASE
			}
		}

		var row entryRow = entryRow{
			date:     end.Format(constants.CARBON_DATE_FORMAT),
			startEnd: start.Format(startEndTimeFormat) + " to " + end.Format(startEndTimeFormat),
			duration: secondsToHuman(util.Round(roundToMinutes, entry.Duration), true),
			project:  entry.Project,
			task:     entry.GetTasksAsString(),
			pushed:   pushed,
			note:     entry.Note,
		}

		dateWidth = widen(dateWidth, row.date)
		startEndWidth = widen(startEndWidth, row.startEnd)
		durationWidth = widen(durationWidth, row.duration)
		pushedWidth = widen(pushedWidth, row.pushed)
		projectWidth = widen(projectWidth, row.project)
		taskWidth = widen(taskWidth, row.task)
		noteWidth = widen(noteWidth, row.note)

		rows = append(rows, row)
	}

	// The date, start/end, duration and pushed columns hold values of a known
	// size, so they are given exactly what they need.  Whatever the terminal
	// has left over is shared between the three variable width columns.
	var columns int = 6
	var fixed int = dateWidth + startEndWidth + durationWidth
	if ticketFound {
		columns = 7
		fixed += pushedWidth
	}

	var available int = terminalWidth - fixed - (columns*columnOverhead + 1)
	var variable []int = []int{projectWidth, taskWidth, noteWidth}

	fitColumns(available, variable, []int{minimumProjectColumnWidth, minimumTaskColumnWidth, minimumNoteColumnWidth})
	projectWidth, taskWidth, noteWidth = variable[0], variable[1], variable[2]

	// Create and configure the table.
	var t table.Writer = table.NewWriter()
	SetReportTableStyle(t)

	var columnConfigs []table.ColumnConfig = []table.ColumnConfig{
		{Number: 1, WidthMin: dateWidth, WidthMax: dateWidth},
		{Number: 2, WidthMin: startEndWidth, WidthMax: startEndWidth},
		{Number: 3, WidthMin: durationWidth, WidthMax: durationWidth},
		{Number: 4, WidthMin: projectWidth, WidthMax: projectWidth, WidthMaxEnforcer: truncateUnicode},
		{Number: 5, WidthMin: taskWidth, WidthMax: taskWidth, WidthMaxEnforcer: truncateUnicode},
	}

	if ticketFound {
		columnConfigs = append(columnConfigs,
			table.ColumnConfig{Number: 6, WidthMin: pushedWidth, WidthMax: pushedWidth},
			table.ColumnConfig{Number: 7, WidthMin: noteWidth, WidthMax: noteWidth, WidthMaxEnforcer: truncateUnicode})

		t.AppendHeader(table.Row{constants.DATE_NORMAL_CASE, constants.START_END_NORMAL_CASE, constants.DURATION_NORMAL_CASE, constants.PROJECT_NORMAL_CASE, constants.TASK_NORMAL_CASE, constants.PUSHED_NORMAL_CASE, constants.NOTE_NORMAL_CASE})
	} else {
		columnConfigs = append(columnConfigs,
			table.ColumnConfig{Number: 6, WidthMin: noteWidth, WidthMax: noteWidth, WidthMaxEnforcer: truncateUnicode})

		t.AppendHeader(table.Row{constants.DATE_NORMAL_CASE, constants.START_END_NORMAL_CASE, constants.DURATION_NORMAL_CASE, constants.PROJECT_NORMAL_CASE, constants.TASK_NORMAL_CASE, constants.NOTE_NORMAL_CASE})
	}

	t.SetColumnConfigs(columnConfigs)

	for _, row := range rows {
		if ticketFound {
			t.AppendRow(table.Row{row.date, row.startEnd, row.duration, row.project, row.task, row.pushed, row.note})
		} else {
			t.AppendRow(table.Row{row.date, row.startEnd, row.duration, row.project, row.task, row.note})
		}
	}

	// Render the table.
	log.Println(t.Render())

	// Export table if needed.
	export("report by entry", t)
}

func reportByLastEntry() {
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var entry models.Entry = db.GetLastEntry()
	var datetime carbon.Carbon = *carbon.Parse(entry.EntryDatetime).SetTimezone(carbon.Local)
	if strings.EqualFold(entry.Project, constants.HELLO) ||
		strings.EqualFold(entry.Project, constants.BREAK) {
		log.Printf("DateTime: %s\n      Project: %s\n    Note: %s\n", datetime.Format("Y-m-d g:i:sa"),
			entry.Project, entry.Note)
	} else {
		log.Printf("DateTime: %s\n Project: %s\n   Tasks: %s\n    Note: %s\n", datetime.Format("Y-m-d g:i:sa"),
			entry.Project, entry.GetTasksAsString(), entry.Note)
	}
}

func reportByProject(entries []models.Entry) {
	log.Printf("\n")
	log.Printf("%s\n", separator(" By Project "))
	log.Printf("\n")

	// Consolidate by project.
	var consolidatedByProject map[string]models.Entry = make(map[string]models.Entry)
	for _, entry := range entries {
		// Check if the project exists in the map or not.
		consolidated, found := consolidatedByProject[entry.Project]
		if found {
			// It already existed, so update it.
			if len(entry.GetTasksAsString()) > 0 {
				consolidated.AddEntryProperty(constants.TASK, entry.GetTasksAsString())
			}
			consolidated.Duration += util.Round(roundToMinutes, entry.Duration)
			consolidatedByProject[entry.Project] = consolidated
		} else {
			var newEntry models.Entry = models.NewEntry(entry.Uid, entry.Project, entry.Note, entry.EntryDatetime)
			newEntry.Duration = util.Round(roundToMinutes, entry.Duration)
			if len(entry.GetTasksAsString()) > 0 {
				newEntry.AddEntryProperty(constants.TASK, entry.GetTasksAsString())
			}
			consolidatedByProject[entry.Project] = newEntry
		}
	}

	// Since maps are not sorted in go... why, I have no idea, you need to first
	// sort the keys and then access the map via those sorted keys.
	var sortedKeys []string = make([]string, 0, len(consolidatedByProject))
	for key := range consolidatedByProject {
		sortedKeys = append(sortedKeys, key)
	}
	sort.SliceStable(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	// Create and configure the table.
	var t table.Writer = table.NewWriter()
	SetReportTableStyle(t)

	t.AppendHeader(table.Row{constants.PROJECT_NORMAL_CASE, constants.TASK_NORMAL_CASE, constants.DURATION_NORMAL_CASE})

	// Add all the consolidated rows to the table.
	for _, i := range sortedKeys {
		var entry models.Entry = consolidatedByProject[i]

		// Skip entries that match constants.HELLO.
		if !strings.EqualFold(entry.Project, constants.HELLO) {
			t.AppendRow(table.Row{entry.Project, entry.GetTasksAsString(), secondsToHuman(entry.Duration, true)})
		}
	}

	// Render the table.
	log.Println(t.Render())

	// Export table if needed.
	export("report by project", t)
}

func reportByTask(entries []models.Entry) {
	log.Printf("\n")
	log.Printf("%s\n", separator(" By Task "))
	log.Printf("\n")

	var consolidateByTask map[string]models.Task = make(map[string]models.Task)
	for _, entry := range entries {
		var task = entry.GetTasksAsString()
		var project = entry.Project
		var key = task + project
		consolidated, found := consolidateByTask[key]
		if found {
			consolidated.Duration += util.Round(roundToMinutes, entry.Duration)
			consolidateByTask[key] = consolidated
		} else {
			var newTask models.Task = models.NewTask(task)
			newTask.Duration = util.Round(roundToMinutes, entry.Duration)
			newTask.AddTaskProperty(constants.PROJECT, entry.Project)
			newTask.AddTaskProperty(constants.TICKET, entry.GetTicketAsString())
			consolidateByTask[key] = newTask
		}
	}

	// Check and see if any entry has a TICKET property.  If so, add it to the table.
	var ticketFound bool = false
	for _, v := range consolidateByTask {
		if !stringUtils.IsBlank(v.GetTicketAsString()) {
			ticketFound = true
			break
		}
	}

	// Create and configure the table.
	var t table.Writer = table.NewWriter()
	SetReportTableStyle(t)

	// If the ticket property was found on any entry, add the URL to the table header.
	if !ticketFound {
		t.AppendHeader(table.Row{constants.TASKS_NORMAL_CASE, constants.PROJECTS_NORMAL_CASE, constants.DURATION_NORMAL_CASE})
	} else {
		t.AppendHeader(table.Row{constants.TASKS_NORMAL_CASE, constants.PROJECTS_NORMAL_CASE, constants.DURATION_NORMAL_CASE, constants.URL_NORMAL_CASE})
	}

	// Since maps are not sorted in go, sort the keys so that the rows are
	// rendered in a stable order instead of a random one.
	var sortedKeys []string = make([]string, 0, len(consolidateByTask))
	for key := range consolidateByTask {
		sortedKeys = append(sortedKeys, key)
	}
	sort.SliceStable(sortedKeys, func(x, y int) bool { return sortedKeys[x] < sortedKeys[y] })

	// Populate the table.
	for _, key := range sortedKeys {
		var v models.Task = consolidateByTask[key]
		if !ticketFound {
			t.AppendRow(table.Row{v.Task, v.GetProjectsAsString(), secondsToHuman(v.Duration, true)})
		} else {
			t.AppendRow(table.Row{v.Task, v.GetProjectsAsString(), secondsToHuman(v.Duration, true), jira.FormatJiraUrl(jira.JiraBrowseTicketUrl, v.GetTicketAsString())})
		}
	}

	// Render the table.
	log.Println(t.Render())

	// Export table if needed.
	export("report by task", t)
}

func reportTotalWorkAndBreakTime(entries []models.Entry) {
	var totalWorkDuration int64 = 0
	var totalBreakDuration int64 = 0

	// Calculate total time worked and total times on break.
	for _, entry := range entries {
		if strings.EqualFold(entry.Project, constants.BREAK) {
			totalBreakDuration += util.Round(roundToMinutes, entry.Duration)
		} else {
			totalWorkDuration += util.Round(roundToMinutes, entry.Duration)
		}
	}

	log.Printf("\n")

	// If we have worked more seconds than are in a day, we need to show hours,
	// minutes, and seconds as well as the human readable form of the duration.
	// By showing the hours, minutes, and seconds, we have a better
	// representation of our duration.  For example... traditionally, a person
	// works 40 hours a week.  If the report tells us we worked 1 day and 3
	// hours... we have to convert that in our heads to 27 hours... But if the
	// report simply did the conversion for us... that is much better.
	if viper.GetBool(constants.SPLIT_WORK_FROM_BREAK_TIME) {
		if totalWorkDuration > constants.SECONDS_PER_DAY {
			log.Printf("Total Working Time: %s (%s)\n", secondsToHuman(totalWorkDuration, true), secondsToHuman(totalWorkDuration, false))
		} else {
			log.Printf("Total Working Time: %s\n", secondsToHuman(totalWorkDuration, true))
		}

		log.Printf("  Total Break Time: %s\n", secondsToHuman(totalBreakDuration, true))
	} else {
		var total = totalWorkDuration + totalBreakDuration
		if totalWorkDuration > constants.SECONDS_PER_DAY {
			log.Printf("Total Time: %s (%s)\n", secondsToHuman(total, true), secondsToHuman(total, false))
		} else {
			log.Printf("Total Time: %s\n", secondsToHuman(total, true))
		}
	}
}

func SetReportTableStyle(t table.Writer) {
	t.SetAllowedRowLength(getTerminalWidth())
	style := table.StyleDefault
	style.Format.Header = text.FormatUpper
	style.Size.WidthMin = terminalWidth
	style.Size.WidthMax = terminalWidth
	t.SetStyle(style)
}

func runReport(cmd *cobra.Command, _ []string) {
	// Save this so we can use it in other methods.
	_cmd = cmd

	// See if the user asked to override round.  If no, use the rounding value
	// from the configuration file.  Otherwise, set the rounding value to 0.
	noRounding, _ := cmd.Flags().GetBool(constants.FLAG_NO_ROUNDING)
	if !noRounding {
		roundToMinutes = viper.GetInt64(constants.ROUND_TO_MINUTES)
	} else {
		roundToMinutes = 0
	}

	currentWeek, _ := cmd.Flags().GetBool(constants.FLAG_CURRENT_WEEK)
	previousWeek, _ := cmd.Flags().GetBool(constants.FLAG_PREVIOUS_WEEK)
	yesterday, _ := cmd.Flags().GetBool(constants.FLAG_YESTERDAY)
	push, _ := cmd.Flags().GetBool(constants.FLAG_PUSH)
	givenDateStr, _ := cmd.Flags().GetString(constants.FLAG_DATE)
	lastEntry, _ := cmd.Flags().GetBool(constants.FLAG_LAST_ENTRY)
	fromDateStr, _ := cmd.Flags().GetString(constants.FLAG_FROM)
	toDateStr, _ := cmd.Flags().GetString(constants.FLAG_TO)
	project, _ := cmd.Flags().GetString(constants.FLAG_PROJECT)

	// If we are supposed to push report items, validate that we first valid push configuration.
	if push {
		pushCredentials = validatePush()
	}

	var now carbon.Carbon = *carbon.Now()
	var start carbon.Carbon = *now.Copy()
	var end carbon.Carbon = *now.Copy()

	if lastEntry {
		reportByLastEntry()
		os.Exit(0)
	} else if stringUtils.IsEmpty(fromDateStr) &&
		stringUtils.IsEmpty(toDateStr) &&
		currentWeek {
		dateRange(&start, &end)
	} else if stringUtils.IsEmpty(fromDateStr) &&
		stringUtils.IsEmpty(toDateStr) &&
		previousWeek {
		start = *start.SubWeek()
		dateRange(&start, &end)
	} else if !stringUtils.IsEmpty(fromDateStr) &&
		!stringUtils.IsEmpty(toDateStr) {
		start = *carbon.Parse(fromDateStr)
		end = *carbon.Parse(toDateStr)
	} else if !stringUtils.IsEmpty(givenDateStr) {
		// Report for given date.
		start = *carbon.Parse(givenDateStr).StartOfDay()
		end = *carbon.Parse(givenDateStr).EndOfDay()
	} else {
		if yesterday {
			// Report for yesterday
			yesterday := carbon.Yesterday()
			start = *yesterday.StartOfDay()
			end = *yesterday.EndOfDay()
		} else {
			// Report for today.
			start = *now.StartOfDay()
			end = *now.EndOfDay()
		}
	}

	if viper.GetBool(constants.DEBUG) {
		log.Printf("\n*****\nStart[%s]\nEnd[%s]\n*****\n", start.ToIso8601String(), end.ToIso8601String())
	}

	var startWeek int = start.WeekOfYear()
	var endWeek int = end.WeekOfYear()

	log.Printf("%s\n", separator(fmt.Sprintf("%s(%d) to %s(%d)", start.ToDateTimeString(), startWeek,
		end.ToDateTimeString(), endWeek)))

	// Get the unique UIDs between the specified start and end dates.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var distinctUIDs []database.DistinctUID = db.GetDistinctUIDs(start, end, project)

	if viper.GetBool(constants.DEBUG) {
		log.Printf("\n*****\nGetDistinctUIDs returned...\n*****\n")
	}

	// Declare the "IN" string used in the db.GetEntries() call.
	var in string = constants.EMPTY

	// Loop through the distinct UIDs and pull out the UID and construct the
	// "in" statement for later use.
	for _, element := range distinctUIDs {
		if viper.GetBool(constants.DEBUG) {
			log.Printf("%d, %s, %s\n", element.Uid, element.Project, element.EntryDatetime)
		}

		if !stringUtils.IsEmpty(in) {
			in = in + ", "
		}

		in = in + strconv.FormatInt(element.Uid, 10)
	}

	// Get all the Entries associated with the list of UIDs.
	var entries []models.Entry = db.GetEntries(in)
	if viper.GetBool(constants.DEBUG) {
		log.Printf("\n*****\nDumping what GetEntries() returned...\n*****\n")
		for _, entry := range entries {
			log.Printf("UID[%d], Project[%s], Note[%#v], EntryDatetime[%s], Properties[%#v]\n",
				entry.Uid, entry.Project, entry.Note, entry.EntryDatetime, entry.GetPropertiesAsString())
		}
	}

	var newEntries []models.Entry

	// Calculate the duration between each UID.
	if viper.GetBool(constants.DEBUG) {
		log.Printf("\n*****\nUpdating entries with durations...\n*****\n")
	}

	for index := range entries {
		// Check to see if the 1st element we have is a HELLO.  If not, we need to adjust
		// accordingly.
		if index == 0 || strings.EqualFold(entries[index].Project, constants.HELLO) {
			var current carbon.Carbon = *carbon.Parse(entries[index].EntryDatetime).SetTimezone(carbon.Local)
			if current.Error != nil {
				log.Fatalf("%s: Unable to parse EntryDateTime. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), current.Error)
				os.Exit(1)
			}

			// Prior is Midnight since this is the 1st record.
			var midnight carbon.Carbon = *current.StartOfDay()
			var entry models.Entry = models.NewEntry(entries[index].Uid, entries[index].Project, entries[index].Note, entries[index].EntryDatetime)
			entry.Properties = entries[index].Properties
			entry.Duration = current.DiffAbsInSeconds(&midnight)
			newEntries = append(newEntries, entry)
		} else {
			var current carbon.Carbon = *carbon.Parse(entries[index].EntryDatetime).SetTimezone(carbon.Local)
			if current.Error != nil {
				log.Fatalf("%s: Unable to parse EntryDateTime. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), current.Error)
				os.Exit(1)
			}

			var prior carbon.Carbon = *carbon.Parse(entries[index-1].EntryDatetime).SetTimezone(carbon.Local)
			if prior.Error != nil {
				log.Fatalf("%s: Unable to parse EntryDateTime. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), prior.Error)
				os.Exit(1)
			}

			// Are the days between the current and prior different?  If they
			// are, that means we went over midnight.
			if !current.IsSameDay(&prior) {
				// Since we have an entry that goes over midnight, we need to
				// create two entries.  One for the time before midnight and one
				// for the time after midnight.
				if viper.GetBool(constants.DEBUG) {
					log.Printf("We went over midnight.\n")
					log.Printf("    current[%s] prior[%s]\n", &current, &prior)
					log.Printf("    prior midnight[%s]\n", prior.EndOfDay())
				}

				// Before midnight.
				var entry models.Entry = models.NewEntry(entries[index].Uid, entries[index].Project, entries[index].Note, prior.EndOfDay().ToRfc3339String())
				entry.Properties = entries[index].Properties
				entry.Duration = prior.EndOfDay().DiffAbsInSeconds(&prior)
				newEntries = append(newEntries, entry)

				// After midnight.
				entry = models.NewEntry(entries[index].Uid, entries[index].Project, entries[index].Note, current.ToRfc3339String())
				entry.Properties = entries[index].Properties
				entry.Duration = current.StartOfDay().DiffAbsInSeconds(&current)
				newEntries = append(newEntries, entry)
			} else {
				var entry models.Entry = models.NewEntry(entries[index].Uid, entries[index].Project, entries[index].Note, entries[index].EntryDatetime)
				entry.Properties = entries[index].Properties
				entry.Duration = current.DiffAbsInSeconds(&prior)
				newEntries = append(newEntries, entry)
			}
		}
	}

	if viper.GetBool(constants.DEBUG) {
		log.Printf("\n*****\nDumping the NEW Entries collection...\n*****\n")
		for index, entry := range newEntries {
			log.Printf("Index[%d] UID[%d], Project[%s], Note[%#v], EntryDatetime[%s], Properties[%#v] Duration[%d or %s]\n",
				index, entry.Uid, entry.Project, entry.Note, entry.EntryDatetime, entry.GetPropertiesAsString(), entry.Duration,
				secondsToHuman(entry.Duration, true))
		}
	}

	var newEntriesWithoutHello []models.Entry
	for index := range newEntries {
		if strings.EqualFold(newEntries[index].Project, constants.HELLO) {
			continue
		} else {
			var entry models.Entry = models.NewEntry(newEntries[index].Uid, newEntries[index].Project, newEntries[index].Note, newEntries[index].EntryDatetime)
			entry.Properties = newEntries[index].Properties
			entry.Duration = newEntries[index].Duration
			newEntriesWithoutHello = append(newEntriesWithoutHello, entry)
		}
	}

	if viper.GetBool(constants.DEBUG) {
		log.Printf("\n*****\nDumping the NEW Entries without HELLOs collection...\n*****\n")
		for index, entry := range newEntriesWithoutHello {
			log.Printf("Index[%d] UID[%d], Project[%s], Note[%#v], EntryDatetime[%s], Properties[%#v] Duration[%d or %s]\n",
				index, entry.Uid, entry.Project, entry.Note, entry.EntryDatetime, entry.GetPropertiesAsString(), entry.Duration,
				secondsToHuman(entry.Duration, true))
		}
	}

	// Check if the user wants 24h formatted time.
	if viper.GetBool(constants.DISPLAY_TIME_IN_24H_FORMAT) {
		startEndTimeFormat = constants.CARBON_START_END_TIME_24H_FORMAT
	}

	// Replace our existing collection of entries with our new collection.
	entries = newEntriesWithoutHello

	// Run each of the reports, if configured to do so.
	reportTotalWorkAndBreakTime(entries)

	if viper.GetBool(constants.REPORT_BY_PROJECT) {
		reportByProject(entries)
	}

	if viper.GetBool(constants.REPORT_BY_TASK) {
		reportByTask(entries)
	}

	if viper.GetBool(constants.REPORT_BY_ENTRY) {
		reportByEntry(entries)
	}

	if viper.GetBool(constants.REPORT_BY_DAY) {
		reportByDay(entries)
	}

	// If the user has asked to push these updates to the server, do so.
	if push {
		pushEntries(db, entries)
	}
}

func validatePush() models.Credentials {
	return rest.ReadCredentials()
}

type HttpRequest struct {
	EntryUid int64
	Request  *http.Request
}

func pushEntries(db *database.Database, entries []models.Entry) {
	var httpRequests []HttpRequest

	// Collect unpushed requests.
	var payloads []jira.JiraRequest = jira.JiraNewRequests(roundToMinutes, entries)

	// Were there any unpushed requests found? If so, process them.
	if len(payloads) > 0 {
		// Transform all the unpushed Jira requests into HTTP requests.
		for _, payload := range payloads {
			request, _ := http.NewRequest("POST", jira.FormatJiraUrl(jira.JiraLogWorkToTicketUrl, payload.Ticket), bytes.NewBuffer(payload.Payload))
			request.Header.Add("Authorization", fmt.Sprintf("Basic %v", rest.BasicAuth(&pushCredentials)))
			request.Header.Add("Content-Type", "application/json")
			var httpRequest HttpRequest = HttpRequest{EntryUid: payload.EntryUid, Request: request}
			httpRequests = append(httpRequests, httpRequest)
		}

		// If in debug, dump all the HTTP requests to the screen.
		if viper.GetBool(constants.DEBUG) {
			log.Printf("\n*****\nDumping all HTTP requests...\n*****\n")
			for _, httpRequest := range httpRequests {
				log.Printf("Entry UID: %d Request: %v\n", httpRequest.EntryUid, httpRequest.Request)
			}
		}

		// Ask the user if they want to push these changes or not.
		yesNo := yesNoPrompt("\nThere are %d unpushed entries. Push them to the server?", len(payloads))
		if yesNo {
			// Yep...
			err := util.RunWithSpinner("Pushing entries", func() error {
				// Attempt to push each entry to the server.
				for _, httpRequest := range httpRequests {
					result, err := rest.HTTPClient.Do(httpRequest.Request)
					if err != nil {
						return fmt.Errorf("failed to send %v: %v", httpRequest, err)
					}
					defer result.Body.Close()
					body, err := io.ReadAll(result.Body)
					if err != nil {
						return fmt.Errorf("failed to read %v", err)
					}

					if viper.GetBool(constants.DEBUG) {
						log.Printf("Jira Server responded: %v\n{%q}\n", result.Status, body)
					}

					// On success, update the entries 'pushed' property.
					if result.StatusCode == 201 {
						db.UpdateEntryPushed(httpRequest.EntryUid)
					} else {
						var entry models.Entry = db.GetEntry(httpRequest.EntryUid)
						return fmt.Errorf("for Entry[%s] Jira Server responded: %v\n{%q}.\nCheck your JIRA configuration for incorrect url, username, and/or api_key.",
							entry.Dump(false, 0), result.Status, body)
					}
				}

				return nil
			})

			// Was any sort of error encountered? I sure hope not.
			if err != nil {
				log.Fatalf("%s: %v\n", color.RedString(constants.FATAL_NORMAL_CASE), err)
				os.Exit(1)
			} else {
				log.Printf("%s\n", color.GreenString("Entries pushed."))
			}
		} else {
			// Nope.
			log.Printf("%s\n", color.YellowString("Entries NOT pushed."))
		}
	}
}

func secondsToHumanFloat(inSeconds float64, hmsOnly bool) (result string) {
	return secondsToHuman(int64(inSeconds), hmsOnly)
}

// Duration components, largest to smallest, used to render a human readable
// duration. A month is 30 days and a year is 365 days.
const (
	secondsPerMinute int64 = 60
	secondsPerHour   int64 = 60 * secondsPerMinute
	secondsPerDay    int64 = 24 * secondsPerHour
	secondsPerWeek   int64 = 7 * secondsPerDay
	secondsPerMonth  int64 = 30 * secondsPerDay
	secondsPerYear   int64 = 365 * secondsPerDay
)

type durationUnit struct {
	seconds      int64
	name         string
	abbreviation string
}

var hmsUnits = []durationUnit{
	{secondsPerHour, "hour", "h"},
	{secondsPerMinute, "minute", "m"},
	{1, "second", "s"},
}

var fullUnits = []durationUnit{
	{secondsPerYear, "year", "y"},
	{secondsPerMonth, "month", "mo"},
	{secondsPerWeek, "week", "w"},
	{secondsPerDay, "day", "d"},
	{secondsPerHour, "hour", "h"},
	{secondsPerMinute, "minute", "m"},
	{1, "second", "s"},
}

func secondsToHuman(inSeconds int64, hmsOnly bool) (result string) {
	var abbreviated bool = viper.GetBool(constants.DISPLAY_HMS_ABBREVIATED)

	// If the duration is zero, this means that the rounded value is less than
	// the "round to minutes" value, simply show a less than message.
	if inSeconds == 0 {
		if roundToMinutes > 0 {
			if abbreviated {
				return fmt.Sprintf("< %dm", roundToMinutes)
			}

			return stringUtils.Trim("< " + plural(int(roundToMinutes), "minute"))
		}

		if abbreviated {
			return "0s"
		}

		return "0 seconds"
	}

	var units []durationUnit = fullUnits
	if hmsOnly {
		units = hmsUnits
	}

	// Walk the units largest to smallest, emitting only the non-zero ones so
	// that a value like 15 minutes renders as "15 minutes" and not as
	// "15 minutes 0 second".
	var builder strings.Builder
	var remaining int64 = inSeconds

	for _, unit := range units {
		var value int64 = remaining / unit.seconds
		remaining = remaining % unit.seconds

		if value == 0 {
			continue
		}

		if abbreviated {
			builder.WriteString(fmt.Sprintf("%d%s ", value, unit.abbreviation))
		} else {
			builder.WriteString(plural(int(value), unit.name))
		}
	}

	return stringUtils.Trim(builder.String())
}

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || width == 0 {
		// Default fallback if terminal size cannot be determined
		return 80
	}
	return width
}

//func truncateWithEllipsis(text string, lineLength int, maxWidth int) string {
//  if lineLength+len(text) <= maxWidth {
//      return text
//  }
//  if maxWidth <= 3 {
//      return text[:maxWidth]
//  }
//
//    var offset int = maxWidth-lineLength-3
//
//    if offset < 0 {
//        offset = 5
//    }
//
//    var result string = constants.EMPTY
//
//    if len(text) >= 5 {
//        result = text[:5]
//    }
//
//  return result + "..."
//}

// truncateUnicode cuts col down to maxLen display columns, appending an
// ellipsis when anything was removed.  Widths are measured in display columns
// rather than bytes so that multi byte runes are never split in half.
func truncateUnicode(col string, maxLen int) string {
	if maxLen <= 0 {
		return constants.EMPTY
	}

	if runewidth.StringWidth(col) <= maxLen {
		return col
	}

	// There is no room for both an ellipsis and any content.
	if maxLen <= 3 {
		return strings.Repeat(".", maxLen)
	}
	truncated := ""
	width := 0
	for _, r := range col {
		rw := runewidth.RuneWidth(r)
		if width+rw > maxLen-3 {
			break
		}
		truncated += string(r)
		width += rw
	}
	return truncated + "..."
}
