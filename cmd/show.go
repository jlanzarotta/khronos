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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"khronos/internal/database"
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
	URL         string `yaml:"url"`
}

func init() {
	showCmd.Flags().BoolVarP(&favorites, constants.FAVORITES, constants.EMPTY, false, "Show favorites")
	showCmd.Flags().BoolVarP(&statistics, constants.STATISTICS, constants.EMPTY, false, "Show statistics")
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
		log.Fatalf("%s: Error unmarshalling configuration file[%s]. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed(), err.Error())
		os.Exit(1)
	}

	log.Printf("Favorites found in configuration file[%s]:\n\n", viper.ConfigFileUsed())

	var urlFound bool = false
	var descriptionFound bool = false
	for _, f := range config.Favorites {
		if len(f.URL) > 0 {
			urlFound = true
		}

		if len(f.Description) > 0 {
			descriptionFound = true
		}
	}

	t.Style().Options.DrawBorder = false
	if descriptionFound && urlFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.DESCRIPTION, constants.URL})
	} else if descriptionFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.DESCRIPTION})
	} else if urlFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.URL})
	} else {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK})
	}

	for i, f := range config.Favorites {
		if descriptionFound && urlFound {
			t.AppendRow(table.Row{i, f.Favorite, f.Description, f.URL})
		} else if descriptionFound {
			t.AppendRow(table.Row{i, f.Favorite, f.Description})
		} else if urlFound {
			t.AppendRow(table.Row{i, f.Favorite, f.URL})
		} else {
			t.AppendRow(table.Row{i, f.Favorite})
		}
	}

	log.Println(t.Render())
}

func showStatistics() {
	db := database.New(viper.GetString(constants.DATABASE_FILE))
	var firstEntry models.Entry = db.GetFirstEntry()
	var lastEntry models.Entry = db.GetLastEntry()
	var count int64 = db.GetCountEntries()

	log.Printf("\n")

	var lastDateTime carbon.Carbon = carbon.Parse(lastEntry.EntryDatetime)
	var firstDateTime carbon.Carbon = carbon.Parse(firstEntry.EntryDatetime)
	var diff int64 = firstDateTime.DiffInSeconds(lastDateTime)

	var t table.Writer = table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.AppendHeader(table.Row{"Statistic", "Value"})
	t.AppendRow(table.Row{"First Entry", firstEntry.Dump(false, 0)})
	t.AppendRow(table.Row{"Last Entry", lastEntry.Dump(false, 0)})
	t.AppendRow(table.Row{"Total Records", count})
	t.AppendRow(table.Row{"Total Duration", secondsToHuman(diff, true)})
	log.Println(t.Render())
}
