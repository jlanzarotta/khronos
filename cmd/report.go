/*
Copyright © 2024 Jeff Lanzarotta
*/
package cmd

import (
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"khronos/constants"
	"khronos/internal/database"
	"khronos/internal/models"

	"golang.org/x/term"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/fatih/color"
	"github.com/golang-module/carbon/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var from string
var to string

var daysOfWeek = map[string]string{}
var roundToMinutes int64

// reportCmd represents the report command.
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a report",
	Long:  `When you need to generate a report, default today, use this command.`,
	Run: func(cmd *cobra.Command, args []string) {
		runReport(cmd, args)
	},
}

func dashes(input string) string {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		log.Fatalf("%s: Error getting terminal dimensions. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	var pad string = strings.Repeat("-", (((width - 2) - len(input)) / 2))
	return (fmt.Sprintf("%s %s %s", pad, input, pad))
}

func dateRange(date carbon.Carbon) (start carbon.Carbon, end carbon.Carbon) {
	start = weekStart(date)
	end = start.AddDays(6).EndOfDay()
	return start, end
}

func init() {
	reportCmd.Flags().BoolP("no-rounding", constants.EMPTY, false, "Reports all durations in their unrounded form.")
	reportCmd.Flags().BoolP("current-week", constants.EMPTY, false, "Report on the current week's entries.")
	reportCmd.Flags().BoolP("previous-week", constants.EMPTY, false, "Report on the previous week's entries.")
	reportCmd.Flags().BoolP("last-entry", constants.EMPTY, false, "Display the last entry's information.")
	reportCmd.Flags().StringVarP(&from, "from", constants.EMPTY, constants.EMPTY, "Specify an inclusive start date to report in "+constants.DATE_FORMAT+" format.")
	reportCmd.Flags().StringVarP(&to, "to", constants.EMPTY, constants.EMPTY, "Specify an inclusive end date to report in "+constants.DATE_FORMAT+" format.  If this is a day of the week, then it is the next occurrence from the start date of the report, including the start date itself.")
	reportCmd.MarkFlagsRequiredTogether("from", "to")
	rootCmd.AddCommand(reportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// reportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// reportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// Populate days of week.
	daysOfWeek[carbon.Sunday] = "Sunday"
	daysOfWeek[carbon.Monday] = "Monday"
	daysOfWeek[carbon.Tuesday] = "Tuesday"
	daysOfWeek[carbon.Wednesday] = "Wednesday"
	daysOfWeek[carbon.Thursday] = "Thursday"
	daysOfWeek[carbon.Friday] = "Friday"
	daysOfWeek[carbon.Saturday] = "Saturday"
}

func parseWeekday(v string) (string, error) {
	if d, ok := daysOfWeek[v]; ok {
		return d, nil
	}

	return "**UNKNOWN**", fmt.Errorf("invalid weekday '%s'", v)
}

func plural(count int, singular string) (result string) {
	if (count == 1) || (count == 0) {
		result = strconv.Itoa(count) + " " + singular + " "
	} else {
		result = strconv.Itoa(count) + " " + singular + "s "
	}

	return
}

func reportByDay(durations map[int64]models.UID, entries []models.Entry) {
	var show_by_day_totals bool = viper.GetBool(constants.SHOW_BY_DAY_TOTALS)
	log.Printf("\n")
	log.Printf("%s\n", dashes(" By Day "))
	log.Printf("\n")

	// Consolidate by day.
	var consolidatedByDay map[string]map[string]models.Entry = make(map[string]map[string]models.Entry)
	for _, e := range entries {
		if strings.EqualFold(e.Project, constants.HELLO) {
			continue
		}

		var task = e.GetTasksAsString()
		consolidatedDay, found := consolidatedByDay[carbon.Parse(e.EntryDatetime).Format(constants.CARBON_DATE_FORMAT)]
		if found {
			consolidatedProject, found := consolidatedDay[e.Project]
			if found {
				if len(task) > 0 {
					consolidatedProject.AddEntryProperty(constants.TASK, task)
				}

				// Add the rounded durations together.
				consolidatedProject.Duration += round(durations[e.Uid].Duration)

				// Replace the consolidated entry.
				consolidatedByDay[carbon.Parse(e.EntryDatetime).Format(constants.CARBON_DATE_FORMAT)][e.Project] = consolidatedProject
			} else {
				var entry models.Entry = models.NewEntry(e.Uid, e.Project, e.Note, e.EntryDatetime)
				entry.Duration = round(durations[e.Uid].Duration)
				if len(task) > 0 {
					entry.AddEntryProperty(constants.TASK, task)
				}

				// Add the new entry.
				consolidatedByDay[carbon.Parse(e.EntryDatetime).Format(constants.CARBON_DATE_FORMAT)][e.Project] = entry
			}
		} else {
			// Since the EntryDatetime was not found, add it.
			var entry models.Entry = models.NewEntry(e.Uid, e.Project, e.Note, e.EntryDatetime)
			entry.Duration = round(durations[e.Uid].Duration)
			if len(task) > 0 {
				entry.AddEntryProperty(constants.TASK, task)
			}

			// Add the new entry.
			var key string = carbon.Parse(entry.EntryDatetime).Format(constants.CARBON_DATE_FORMAT)
			consolidatedByDay[key] = make(map[string]models.Entry)
			consolidatedByDay[key][entry.Project] = entry
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
	t.Style().Options.DrawBorder = false
	t.AppendHeader(table.Row{constants.DATE_NORMAL_CASE, constants.PROJECT_NORMAL_CASE, constants.TASKS_NORMAL_CASE, constants.DURATION_NORMAL_CASE})

	// Add each row to the table.
	for _, i := range sortedKeys {
		var day map[string]models.Entry = consolidatedByDay[i]
		var totalPerDay int64 = 0

		for p, v := range day {
			t.AppendRow(table.Row{i, p, v.GetTasksAsString(), secondsToHuman(v.Duration, true)})
			totalPerDay += round(v.Duration)
		}

		if show_by_day_totals {
			t.AppendSeparator()
			t.AppendRow(table.Row{"", "", constants.TOTAL, secondsToHuman(totalPerDay, true)})
			t.AppendSeparator()
		}
	}

	// Render the table.
	log.Println(t.Render())
}

func reportByEntry(durations map[int64]models.UID, entries []models.Entry) {
	log.Printf("\n")
	log.Printf("%s\n", dashes(" By Entry "))
	log.Printf("\n")

	// Consolidate
	var consolidatedByUid map[int64]models.Entry = make(map[int64]models.Entry)
	for _, e := range entries {
		var task = e.GetTasksAsString()
		// Check if the project exists in the map or not.
		consolidated, found := consolidatedByUid[e.Uid]
		if found {
			if len(task) > 0 {
				consolidated.AddEntryProperty(constants.TASK, task)
			}

			// Add the consolidated object to the collection.
			consolidatedByUid[e.Uid] = consolidated
		} else {
			var entry models.Entry = models.NewEntry(e.Uid, e.Project, e.Note, e.EntryDatetime)
			entry.Duration = durations[e.Uid].Duration
			if len(task) > 0 {
				entry.AddEntryProperty(constants.TASK, task)
			}
			consolidatedByUid[e.Uid] = entry
		}
	}

	// Since maps are not sorter in go... why, I have no idea, you need to first
	// sort the keys and then access the map via those sorted keys.
	var sortedKeys []int64 = make([]int64, 0, len(consolidatedByUid))
	for key := range consolidatedByUid {
		sortedKeys = append(sortedKeys, key)
	}
	sort.SliceStable(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

	// Create and configure the table.
	var t table.Writer = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.AppendHeader(table.Row{constants.DATE_NORMAL_CASE, constants.START_END_NORMAL_CASE, constants.DURATION_NORMAL_CASE, constants.PROJECT_NORMAL_CASE, constants.TASK_NORMAL_CASE, constants.NOTE_NORMAL_CASE})

	// Add all the consolidated rows to the table.
	for _, i := range sortedKeys {
		var entry models.Entry = consolidatedByUid[i]

		// Skip entries that match constants.HELLO.
		if !strings.EqualFold(entry.Project, constants.HELLO) {
			t.AppendRow(table.Row{
				carbon.Parse(entry.EntryDatetime).Format(constants.CARBON_DATE_FORMAT),
				carbon.Parse(entry.EntryDatetime).SubSeconds(int(entry.Duration)).Format(constants.CARBON_START_END_TIME_FORMAT) + " to " + carbon.Parse(entry.EntryDatetime).Format(constants.CARBON_START_END_TIME_FORMAT),
				secondsToHuman(round(entry.Duration), true),
				entry.Project,
				entry.GetTasksAsString(),
				entry.Note})
		}
	}

	// Render the table.
	log.Println(t.Render())
}

func reportByLastEntry() {
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var entry models.Entry = db.GetLastEntry()
	if strings.EqualFold(entry.Project, constants.HELLO) ||
		strings.EqualFold(entry.Project, constants.BREAK) {
		log.Printf("DateTime: %s\n      Project: %s\n    Note: %s\n", carbon.Parse(entry.EntryDatetime).Format("Y-m-d g:i:sa"), entry.Project, entry.Note)
	} else {
		log.Printf("DateTime: %s\n Project: %s\n   Tasks: %s\n    Note: %s\n", carbon.Parse(entry.EntryDatetime).Format("Y-m-d g:i:sa"), entry.Project, entry.GetTasksAsString(), entry.Note)
	}
}

func reportByProject(durations map[int64]models.UID, entries []models.Entry) {
	log.Printf("\n")
	log.Printf("%s\n", dashes(" By Project "))
	log.Printf("\n")

	// Consolidate by project.
	var consolidatedByProject map[string]models.Entry = make(map[string]models.Entry)
	for _, e := range entries {
		// Check if the project exists in the map or not.
		consolidated, found := consolidatedByProject[e.Project]
		if found {
			if len(e.GetTasksAsString()) > 0 {
				consolidated.AddEntryProperty(constants.TASK, e.GetTasksAsString())
			}

			// If the Uid changes, add the new duration.
			if consolidated.Uid != e.Uid {
				consolidated.Uid = e.Uid
				consolidated.Duration += round(durations[e.Uid].Duration)
			}

			// Add the consolidated object to the collection.
			consolidatedByProject[e.Project] = consolidated
		} else {
			var entry models.Entry = models.NewEntry(e.Uid, e.Project, e.Note, e.EntryDatetime)
			entry.Duration = round(durations[e.Uid].Duration)
			if len(e.GetTasksAsString()) > 0 {
				entry.AddEntryProperty(constants.TASK, e.GetTasksAsString())
			}
			consolidatedByProject[e.Project] = entry
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
	t.Style().Options.DrawBorder = false
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
}

func reportByTask(durations map[int64]models.UID, entries []models.Entry) {
	log.Printf("\n")
	log.Printf("%s\n", dashes(" By Task "))
	log.Printf("\n")

	var consolidateByTask map[string]models.Task = make(map[string]models.Task)
	for _, e := range entries {
		if strings.EqualFold(e.Project, constants.HELLO) {
			continue
		} else {
			var t = e.GetTasksAsString()
			consolidated, found := consolidateByTask[t]
			if found {
				consolidated.Duration += round(durations[e.Uid].Duration)
				consolidateByTask[t] = consolidated
			} else {
				var task models.Task = models.NewTask(t)
				task.Duration = round(durations[e.Uid].Duration)
				task.AddTaskProperty(constants.PROJECT, e.Project)
				task.AddTaskProperty(constants.URL, e.GetUrlAsString())
				consolidateByTask[t] = task
			}
		}
	}

	// Check and see if any entry has a URL property.  If so, add it to the table.
	var urlFound bool = false
	for _, v := range consolidateByTask {
		if len(v.GetUrlAsString()) > 0 {
			urlFound = true
			break
		}
	}

	// Create and configure the table.
	var t table.Writer = table.NewWriter()
	t.Style().Options.DrawBorder = false
	if !urlFound {
		t.AppendHeader(table.Row{constants.TASKS_NORMAL_CASE, constants.PROJECTS_NORMAL_CASE, constants.DURATION_NORMAL_CASE})
	} else {
		t.AppendHeader(table.Row{constants.TASKS_NORMAL_CASE, constants.PROJECTS_NORMAL_CASE, constants.DURATION_NORMAL_CASE, constants.URL_NORMAL_CASE})
	}

	// Populate the table.
	for _, v := range consolidateByTask {
		if !urlFound {
			t.AppendRow(table.Row{v.Task, v.GetProjectsAsString(), secondsToHuman(v.Duration, true)})
		} else {
			t.AppendRow(table.Row{v.Task, v.GetProjectsAsString(), secondsToHuman(v.Duration, true), v.GetUrlAsString()})
		}
	}

	// Render the table.
	log.Println(t.Render())
}

func reportTotalWorkAndBreakTime(durations map[int64]models.UID, entries []models.Entry) {
	var totalWorkDuration int64 = 0
	var totalBreakDuration int64 = 0

	// Calculate total time worked and total times on break.
	for _, e := range entries {
		// Skip HELLOs.
		if strings.EqualFold(e.Project, constants.HELLO) {
			continue
		} else if strings.EqualFold(e.Project, constants.BREAK) {
			totalBreakDuration += round(durations[e.Uid].Duration)
		} else {
			totalWorkDuration += round(durations[e.Uid].Duration)
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

func round(durationInSeconds int64) (result int64) {
	var seconds int64 = durationInSeconds

	if roundToMinutes > 0 {
		var remainder int64 = seconds % (roundToMinutes * 60)
		seconds -= remainder
		if remainder/6000 >= 8 {
			// Round up since we are over the threshold of precision.
			seconds = seconds + roundToMinutes*60
		}
	}

	return (seconds)
}

func runReport(cmd *cobra.Command, _ []string) {
	var start carbon.Carbon
	var end carbon.Carbon

	// See if the user asked to override round.  If no, use the rounding value
	// from the configuration file.  Otherwise, set the rounding value to 0.
	noRounding, _ := cmd.Flags().GetBool("no-rounding")
	if !noRounding {
		roundToMinutes = viper.GetInt64(constants.ROUND_TO_MINUTES)
	} else {
		roundToMinutes = 0
	}

	currentWeek, _ := cmd.Flags().GetBool("current-week")
	previousWeek, _ := cmd.Flags().GetBool("previous-week")
	lastEntry, _ := cmd.Flags().GetBool("last-entry")
	fromDateStr, _ := cmd.Flags().GetString("from")
	toDateStr, _ := cmd.Flags().GetString("to")

	var reportNow = carbon.Now()

	if lastEntry {
		reportByLastEntry()
		os.Exit(0)
	} else if stringUtils.IsEmpty(fromDateStr) &&
		stringUtils.IsEmpty(toDateStr) &&
		currentWeek {
		start, end = dateRange(reportNow)
	} else if stringUtils.IsEmpty(fromDateStr) &&
		stringUtils.IsEmpty(toDateStr) &&
		previousWeek {
		start, end = dateRange(reportNow.SubWeek())
	} else if !stringUtils.IsEmpty(fromDateStr) &&
		!stringUtils.IsEmpty(toDateStr) {
		start = carbon.Parse(fromDateStr)
		end = carbon.Parse(toDateStr)
	} else {
		// Report for today.
		start = carbon.Now().StartOfDay()
		end = carbon.Now().EndOfDay()
	}

	var startWeek int = start.WeekOfYear()
	var endWeek int = end.WeekOfYear()

	log.Printf("%s\n", dashes(fmt.Sprintf("%s(%d) to %s(%d)",
		start, startWeek, end, endWeek)))

	// Get the unique UIDs between the specified start and end dates.
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var distinctUIDs []database.DistinctUID = db.GetDistinctUIDs(start, end)

	if viper.GetBool("debug") {
		log.Printf("\n*****\nGetDistinctUIDs returned...\n*****\n")
	}

	// Declare the "IN" string used in the db.GetEntries() call.
	var in string = constants.EMPTY

	// Loop through the distinct UIDs and pull out the UID and construct the
	// "in" statement for later use.
	for _, element := range distinctUIDs {
		if viper.GetBool("debug") {
			log.Printf("%d, %s, %s\n", element.Uid, element.Project, element.EntryDatetime)
		}

		if !stringUtils.IsEmpty(in) {
			in = in + ", "
		}

		in = in + strconv.FormatInt(element.Uid, 10)
	}

	// Calculate the duration between each UID.
	if viper.GetBool("debug") {
		log.Printf("\n*****\nCalculating Durations...\n*****\n")
	}

	var durations map[int64]models.UID = make(map[int64]models.UID)
	for i := range distinctUIDs {
		// Check to see if the 1st element we have is a HELLO.  If not, we need to adjust
		// accordingly.
		if i == 0 || strings.EqualFold(distinctUIDs[i].Project, constants.HELLO) {
			var current carbon.Carbon = carbon.Parse(distinctUIDs[i].EntryDatetime)
			if current.Error != nil {
				log.Fatalf("%s: Unable to parse EntryDateTime. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), current.Error)
				os.Exit(1)
			}

			// Prior is Midnight since this is the 1st record.
			var midnight carbon.Carbon = current.StartOfDay()
			var uid models.UID = models.NewUID(distinctUIDs[i].Uid, distinctUIDs[i].EntryDatetime, current.DiffAbsInSeconds(midnight))
			durations[distinctUIDs[i].Uid] = uid
		} else {
			var current carbon.Carbon = carbon.Parse(distinctUIDs[i].EntryDatetime)
			if current.Error != nil {
				log.Fatalf("%s: Unable to parse EntryDateTime. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), current.Error)
				os.Exit(1)
			}

			var prior carbon.Carbon = carbon.Parse(distinctUIDs[i-1].EntryDatetime)
			if prior.Error != nil {
				log.Fatalf("%s: Unable to parse EntryDateTime. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), prior.Error)
				os.Exit(1)
			}

			var uid models.UID = models.NewUID(distinctUIDs[i].Uid, distinctUIDs[i].EntryDatetime, current.DiffAbsInSeconds(prior))
			durations[distinctUIDs[i].Uid] = uid
		}
	}

	// If requested, dump all the data with the newly rounded durations.
	if viper.GetBool("debug") {
		log.Printf("\n*****\nDumping newly calculated duration...\n*****\n")

		// Since maps are not sorter in go... why, I have no idea, you need to first
		// sort the keys and then access the map via those sorted keys.
		var sortedKeys []int64 = make([]int64, 0, len(durations))
		for key := range durations {
			sortedKeys = append(sortedKeys, key)
		}

		// Sort the keys.
		sort.SliceStable(sortedKeys, func(i, j int) bool { return sortedKeys[i] < sortedKeys[j] })

		for _, i := range sortedKeys {
			log.Printf("Key[%d] Uid[%d] EntryDatetime[%s] Duration[%d or %s]\n",
				i, durations[i].Uid, durations[i].EntryDatetime, durations[i].Duration,
				secondsToHuman(durations[i].Duration, true))
		}
	}

	// Get all the Entries associated with the list of UIDs.
	var entries []models.Entry = db.GetEntries(in)
	if viper.GetBool("debug") {
		log.Printf("\n*****\nDumping what GetEntries() returned...\n*****\n")
		for _, element := range entries {
			log.Printf("%d, %s, %#v, %s, %#v\n",
				element.Uid, element.Project, element.Note, element.EntryDatetime,
				element.GetPropertiesAsString())
		}
	}

	// Run each of the reports, if configured to do so.
	reportTotalWorkAndBreakTime(durations, entries)

	if viper.GetBool(constants.REPORT_BY_PROJECT) {
		reportByProject(durations, entries)
	}

	if viper.GetBool(constants.REPORT_BY_TASK) {
		reportByTask(durations, entries)
	}

	if viper.GetBool(constants.REPORT_BY_ENTRY) {
		reportByEntry(durations, entries)
	}

	if viper.GetBool(constants.REPORT_BY_DAY) {
		reportByDay(durations, entries)
	}
}

func secondsToHuman(inSeconds int64, hmsOnly bool) (result string) {
	// If the duration is zero, this means than the rounded value is less than
	// the "round to minutes" value, simply show a less than message.
	if inSeconds == 0 {
		result = "< " + plural(int(roundToMinutes), "minute")
	} else {
		if (hmsOnly) {
			hours := inSeconds / 3600
			inSeconds = inSeconds % 3600
			minutes := inSeconds / 60
			seconds := inSeconds % 60

			if hours > 0 {
				result = plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
			} else if minutes > 0 {
				result = plural(int(minutes), "minute") + plural(int(seconds), "second")
			} else {
				result = plural(int(seconds), "second")
			}
		} else {
			// The duration is greater than zero, so process it.
			years := math.Floor(float64(inSeconds) / 60 / 60 / 24 / 7 / 30 / 12)
			seconds := inSeconds % (60 * 60 * 24 * 7 * 30 * 12)
			months := math.Floor(float64(seconds) / 60 / 60 / 24 / 7 / 30)
			seconds = inSeconds % (60 * 60 * 24 * 7 * 30)
			weeks := math.Floor(float64(seconds) / 60 / 60 / 24 / 7)
			seconds = inSeconds % (60 * 60 * 24 * 7)
			days := math.Floor(float64(seconds) / 60 / 60 / 24)
			seconds = inSeconds % (60 * 60 * 24)
			hours := math.Floor(float64(seconds) / 60 / 60)
			seconds = inSeconds % (60 * 60)
			minutes := math.Floor(float64(seconds) / 60)
			seconds = inSeconds % 60

			if years > 0 {
				result = plural(int(years), "year") + plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
			} else if months > 0 {
				result = plural(int(months), "month") + plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
			} else if weeks > 0 {
				result = plural(int(weeks), "week") + plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
			} else if days > 0 {
				result = plural(int(days), "day") + plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
			} else if hours > 0 {
				result = plural(int(hours), "hour") + plural(int(minutes), "minute") + plural(int(seconds), "second")
			} else if minutes > 0 {
				result = plural(int(minutes), "minute") + plural(int(seconds), "second")
			} else {
				result = plural(int(seconds), "second")
			}
		}
	}

	return stringUtils.Trim(result)
}

func weekStart(date carbon.Carbon) carbon.Carbon {
	dayOfWeek, err := parseWeekday(viper.GetString(constants.WEEK_START))
	if err != nil {
		log.Fatalf("%s: %s is an invalid day of week.  Please correct your configuration.\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.GetString(constants.WEEK_START))
		os.Exit(1)
	}

	return date.SetWeekStartsAt(dayOfWeek).StartOfWeek()
}
