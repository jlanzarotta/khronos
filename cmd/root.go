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
	"errors"
	"khronos/constants"
	"khronos/internal/database"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var note string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "khronos",
	Short: constants.ROOT_SHORT_DESCRIPTION,
	Long:  constants.ROOT_LONG_DESCRIPTION,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	cobra.AddTemplateFunc("StyleHeading", color.New(color.FgGreen).SprintFunc())
	usageTemplate := rootCmd.UsageTemplate()
	usageTemplate = strings.NewReplacer(
		`Usage:`, `{{StyleHeading "Usage:"}}`,
		`Aliases:`, `{{StyleHeading "Aliases:"}}`,
		`Available Commands:`, `{{StyleHeading "Available Commands:"}}`,
		`Global Flags:`, `{{StyleHeading "Global Flags:"}}`,
		// The following one steps on "Global Flags:"
		`Flags:`, `{{StyleHeading "Flags:"}}`,
	).Replace(usageTemplate)
	re := regexp.MustCompile(`(?m)^Flags:\s*$`)
	usageTemplate = re.ReplaceAllLiteralString(usageTemplate, `{{StyleHeading "Flags:"}}`)
	rootCmd.SetUsageTemplate(usageTemplate)
	rootCmd.SetOut(color.Output)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", constants.EMPTY, "config file (default is $HOME/.khronos.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.PersistentFlags().Bool("help", false, constants.HELP_SHORT_DESCRIPTION)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Find home directory.
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	if cfgFile != constants.EMPTY {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Add the XDG_CONFIG_HOME directory to the search path, if configured.
		xdgConfigHome, found := os.LookupEnv("XDG_CONFIG_HOME")
		if found {
			xdgConfigPath := filepath.Join(xdgConfigHome, strings.ToLower(constants.APPLICATION_NAME))
			_, err := os.Stat(xdgConfigPath)
			if os.IsNotExist(err) {
				// Do nothing.
			} else {
				viper.AddConfigPath(xdgConfigPath)
			}
		}

		// Add the user's home directory to the search path.
		viper.AddConfigPath(home)

		// Add the Khronos configuration file and extension type.
		viper.SetConfigType("yaml")
		viper.SetConfigName(".khronos")
	}

	// Read in environment variables that match.
	viper.AutomaticEnv()

	// Set default database.
	viper.SetDefault(constants.DATABASE_FILE, filepath.Join(home, ".khronos.db"))

	// Set debug to false.
	viper.SetDefault(constants.DEBUG, false)

	// Should a daily total be shown for each day when rendering the "by day"
	// report.
	viper.SetDefault(constants.DISPLAY_BY_DAY_TOTALS, true)

	// Set 24h time format to false.
	viper.SetDefault(constants.DISPLAY_TIME_IN_24H_FORMAT, false)

	// Set each of the reports to true.
	viper.SetDefault(constants.REPORT_BY_PROJECT, true)
	viper.SetDefault(constants.REPORT_BY_TASK, true)
	viper.SetDefault(constants.REPORT_BY_ENTRY, true)
	viper.SetDefault(constants.REPORT_BY_DAY, true)

	// Require a note.
	viper.SetDefault(constants.REQUIRE_NOTE, false)

	// Round to 15 minute intervals by default.
	viper.SetDefault(constants.ROUND_TO_MINUTES, 15)

	// Set flag indicating if work and break time should be spit into separate values during reports.
	viper.SetDefault(constants.SPLIT_WORK_FROM_BREAK_TIME, false)

	// Set day of the week when determining start of the week.
	viper.SetDefault(constants.WEEK_START, "Sunday")

	// Read the configuration file.
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// No config file, just use defaults.
			viper.SafeWriteConfig()
			//writeDefaultFavorites(home)
			writeDefaultFavorites(viper.ConfigFileUsed())
			log.Printf("%s: Unable to load config file, using/writing default values to [%s].\n\n",
				color.HiBlueString(constants.INFO_NORMAL_CASE), viper.ConfigFileUsed())
		} else {
			log.Fatalf("%s: Error reading config file: %s\n",
				color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}
	}

	// Dump our some debug information.
	if viper.GetBool(constants.DEBUG) {
		log.Printf("%s = [%s]\n", constants.WEEK_START, viper.GetString(constants.WEEK_START))
		log.Printf("%s = [%d]\n", constants.ROUND_TO_MINUTES, viper.GetInt64(constants.ROUND_TO_MINUTES))
		log.Printf("%s = [%v]\n", constants.SPLIT_WORK_FROM_BREAK_TIME, viper.GetBool(constants.SPLIT_WORK_FROM_BREAK_TIME))
	}

	// Check if the database exists or not.  If it does not, create it.
	_, err = os.Stat(viper.GetString(constants.DATABASE_FILE))
	if errors.Is(err, os.ErrNotExist) {
		log.Printf("Database[%s] does not exist, creating...", viper.GetString(constants.DATABASE_FILE))

		var filename string = viper.GetString(constants.DATABASE_FILE)
		os.Create(filename)

		db := database.New(viper.GetString(constants.DATABASE_FILE))
		db.Create()
	}
}

func writeDefaultFavorites(home string) {
	// Populate the configuration file path and name.  We need to play some
	// games here so that viper has a configuration file so we can append to it.
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".khronos")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("%s: Error reading config file: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	// Open our configuration file.
	f, err := os.OpenFile(viper.ConfigFileUsed(), os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("%s: Unable to write favorites to configuration file[%s].\n", color.RedString(constants.FATAL_NORMAL_CASE), viper.ConfigFileUsed())
		os.Exit(1)
	}

	// Remember to close the file.
	defer f.Close()

	var lines = []string{
		"favorites:",
		"  - favorite: general+training",
		"  - favorite: general+product development",
		"  - favorite: general+personal time",
		"  - favorite: general+holiday",
		"  - favorite: general+vacation/PTO/Comp",
	}

	// Write our default favorites to the configuration file.
	for _, line := range lines {
		_, err := f.WriteString(line + "\n")
		if err != nil {
			log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		}
	}
}
