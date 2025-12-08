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
	"log"
	"os"

	"github.com/dromara/carbon/v2"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"khronos/internal/database"
	"khronos/internal/jira"
	"khronos/internal/models"
)

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: constants.SHOW_SHORT_DESCRIPTION,
	Long:  constants.SHOW_LONG_DESCRIPTION,
	Run: func(cmd *cobra.Command, args []string) {
		runShow(cmd, args)
	},
}

var favorites bool
var statistics bool
var unpushed bool

type Configuration struct {
	DatabaseFilename string     `yaml:"database_file"`
	WeekStart        string     `yaml:"week_start"`
	RoundToMinutes   int        `yaml:"round_to_minutes"`
	Debug            bool       `yaml:"debug"`
	Favorites        []Favorite `yaml:"favorites"`
}

type Favorite struct {
	Favorite    string `yaml:"favorite"`
	Description string `yaml:"description"`
	Ticket      string `yaml:"ticket"`
	RequireNote bool   `default:"false" yaml:"require_note"`
}

func init() {
	showCmd.Flags().BoolVarP(&favorites, constants.FAVORITES, constants.EMPTY, false, "Show favorites")
	showCmd.Flags().BoolVarP(&statistics, constants.STATISTICS, constants.EMPTY, false, "Show statistics")
	showCmd.Flags().BoolVarP(&unpushed, constants.UNPUSHED, constants.EMPTY, false, "Show unpushed entries")
	rootCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runShow(cmd *cobra.Command, _ []string) {
	// Get the --favorites flag.
	favorites, _ := cmd.Flags().GetBool(constants.FAVORITES)
	statistics, _ := cmd.Flags().GetBool(constants.STATISTICS)

	if favorites {
		showFavorites()
	}

	if statistics {
		showStatistics()
	}

	if unpushed {
		showUnpushedEntries()
	}
}

func showFavorites() {
	data, err := os.ReadFile(viper.ConfigFileUsed())
	if err != nil {
		log.Fatalf("%s: Error reading configuration file[%s]. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	var config Configuration
	var t table.Writer = table.NewWriter()

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("%s: Error unmarshaling configuration file[%s]. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	log.Printf("Favorites found in configuration file[%s]:\n\n", viper.ConfigFileUsed())

	var ticketFound bool = false
	var descriptionFound bool = false
	for _, f := range config.Favorites {
		if len(f.Ticket) > 0 {
			ticketFound = true
		}

		if len(f.Description) > 0 {
			descriptionFound = true
		}
	}

	t.SetStyle(table.Style{
		Name: "ShowFavorites",
		Box: table.BoxStyle{
			MiddleHorizontal: "-",
			MiddleSeparator:  "+",
			MiddleVertical:   "|",
			PaddingLeft:      " ",
			PaddingRight:     " ",
		},
		Color: table.ColorOptions{
			IndexColumn:  text.Colors{text.BgCyan, text.FgBlack},
			Row:          text.Colors{text.BgBlack, text.FgWhite},
			RowAlternate: text.Colors{text.BgBlack, text.FgHiWhite},
		},
		Format: table.FormatOptions{
			Header: text.FormatUpper,
			Row:    text.FormatDefault,
		},
		Options: table.Options{
			DrawBorder:      false,
			SeparateColumns: true,
			SeparateFooter:  false,
			SeparateHeader:  true,
			SeparateRows:    false,
		},
	})

	if descriptionFound && ticketFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.DESCRIPTION, constants.URL, constants.REQUIRE_NOTE_WITH_ASTERISK})
	} else if descriptionFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.DESCRIPTION, constants.REQUIRE_NOTE_WITH_ASTERISK})
	} else if ticketFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.URL, constants.REQUIRE_NOTE_WITH_ASTERISK})
	} else {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.REQUIRE_NOTE_WITH_ASTERISK})
	}

	for i, f := range config.Favorites {
		if descriptionFound && ticketFound {
			t.AppendRow(table.Row{i, f.Favorite, f.Description, jira.FormatJiraUrl(jira.JiraBrowseTicketUrl, f.Ticket), f.RequireNote})
		} else if descriptionFound {
			t.AppendRow(table.Row{i, f.Favorite, f.Description, f.RequireNote})
		} else if ticketFound {
			t.AppendRow(table.Row{i, f.Favorite, jira.FormatJiraUrl(jira.JiraBrowseTicketUrl, f.Ticket), f.RequireNote})
		} else {
			t.AppendRow(table.Row{i, f.Favorite, f.RequireNote})
		}
	}

	log.Println(t.Render())
	log.Printf("\n%s\n", constants.MAY_BE_OVERRIDDEN_BY_GLOBAL_CONFIGURATION_SETTING)
}

func showStatistics() {
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var firstEntry models.Entry = db.GetFirstEntry()
	var lastEntry models.Entry = db.GetLastEntry()
	var count int64 = db.GetCountEntries()

	log.Printf("\n")

	var lastDateTime carbon.Carbon = *carbon.Parse(lastEntry.EntryDatetime)
	var firstDateTime carbon.Carbon = *carbon.Parse(firstEntry.EntryDatetime)
	var diff int64 = firstDateTime.DiffInSeconds(&lastDateTime)

	var t table.Writer = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.AppendHeader(table.Row{"Statistic", "Value"})
	t.AppendRow(table.Row{"First Entry", firstEntry.Dump(false, 0)})
	t.AppendRow(table.Row{"Last Entry", lastEntry.Dump(false, 0)})
	t.AppendRow(table.Row{"Total Records", count})
	t.AppendRow(table.Row{"Total Duration", secondsToHuman(diff, true)})
	log.Println(t.Render())
}

func showUnpushedEntries() {
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var entries []models.Entry = db.GetUnpushedEntries()
	if viper.GetBool(constants.DEBUG) {
		log.Printf("\n*****\nDumping what GetUnpushedEntries() returned...\n*****\n")
		for _, entry := range entries {
			log.Printf("UID[%d], Project[%s], Note[%#v], EntryDatetime[%s], Properties[%#v]\n",
				entry.Uid, entry.Project, entry.Note, entry.EntryDatetime, entry.GetPropertiesAsString())
		}
	}

	// Create and configure the table.
	var t table.Writer = table.NewWriter()
	SetReportTableStyle(t)

	t.AppendHeader(table.Row{constants.DATE_NORMAL_CASE, constants.PROJECT_NORMAL_CASE, constants.TASK_NORMAL_CASE, constants.NOTE_NORMAL_CASE})

	for _, entry := range entries {
		var entryDatetime carbon.Carbon = *carbon.Parse(entry.EntryDatetime).SetTimezone(carbon.Local)
		t.AppendRow(table.Row{
			entryDatetime.Format(constants.CARBON_DATE_FORMAT),
			entry.Project,
			entry.GetTasksAsString(),
			entry.Note})
	}

	// Render the table.
	log.Println(t.Render())
}
