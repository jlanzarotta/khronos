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
	"bytes"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"
	"khronos/constants"
	"khronos/internal/database"

	"github.com/fatih/color"
	"github.com/dromara/carbon/v2"
	"github.com/inancgumus/screen"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var nukeCmd = &cobra.Command{
	Use:   "nuke",
	Short: constants.NUKE_SHORT_DESCRIPTION,
	Long:  constants.NUKE_LONG_DESCRIPTION,
	Run: func(cmd *cobra.Command, args []string) {
		runNuke(cmd, args)
	},
}

func init() {
	nukeCmd.Flags().BoolP(constants.NUKE_ALL, constants.EMPTY, false, constants.NUKE_ALL_DESCRIPTION)
	nukeCmd.Flags().BoolP(constants.PRIOR_YEARS, constants.EMPTY, false, constants.PRIOR_YEARS_DESCRIPTION)
	nukeCmd.Flags().BoolP(constants.DRY_RUN, constants.EMPTY, false, constants.DRY_RUN_DESCRIPTION)
	nukeCmd.Flags().BoolP(constants.ARCHIVE, constants.EMPTY, false, constants.ARCHIVE_DESCRIPTION)
	nukeCmd.Flags().BoolP(constants.COMPRESS, constants.EMPTY, false, constants.COMPRESS_DESCRIPTION)
	rootCmd.AddCommand(nukeCmd)
}

func runNuke(cmd *cobra.Command, _ []string) {
	all, _ := cmd.Flags().GetBool(constants.NUKE_ALL)
	priorYears, _ := cmd.Flags().GetBool(constants.PRIOR_YEARS)
	dryRun, _ := cmd.Flags().GetBool(constants.DRY_RUN)
	archive, _ := cmd.Flags().GetBool(constants.ARCHIVE)
	compress, _ := cmd.Flags().GetBool(constants.COMPRESS)

	if all {
		yesNo := yesNoPrompt("Are you sure you want to nuke ALL the entries from your database?")
		if yesNo {
			yesNo = yesNoPrompt("WARNING: Are you REALLY sure you want to nuke ALL the entries from your database?")
			if yesNo {
				yesNo = yesNoPrompt("LAST WARNING: Are you REALLY REALLY sure you want to nuke ALL the entries from your database?")
				if yesNo {
					// Yes was enter, so nuke ALL entries.
					db := database.New(viper.GetString(constants.DATABASE_FILE))
					var count = db.NukeAllEntries(dryRun, archive, compress)
					showExplosion()
					if dryRun {
						log.Printf("%s\n", color.HiBlueString("All %d entries would have been nuked.", count))
					} else {
						log.Printf("%s\n", color.GreenString("All entries nuked."))
					}
				} else {
					log.Printf("%s\n", color.YellowString("Nothing nuked."))
				}
			} else {
				log.Printf("%s\n", color.YellowString("Nothing nuked."))
			}
		} else {
			log.Printf("%s\n", color.YellowString("Nothing nuked."))
		}
	} else if priorYears {
		var year int = carbon.Now().Year()
		var prompt = fmt.Sprintf("Are you sure you want to nuke all entries prior to %d from the database?", year)
		yesNo := yesNoPrompt(prompt)
		if yesNo {
			prompt = fmt.Sprintf("WARNING: Are you REALLY sure you want to nuke all entries prior to %d from the database?", year)
			yesNo = yesNoPrompt(prompt)
			if yesNo {
				prompt = fmt.Sprintf("LAST WARNING: Are you REALLY REALLY sure you want to nuke all entries prior to %d from the database?", year)
				yesNo = yesNoPrompt(prompt)
				if yesNo {
					db := database.New(viper.GetString(constants.DATABASE_FILE))
					var count = db.NukePriorYearsEntries(dryRun, year, archive, compress)
					showExplosion()
					if dryRun {
						log.Printf("%s\n", color.YellowString("All %d entries prior to %d would have been nuked.\n", count, year))
					} else {
						log.Printf("%s\n", color.GreenString("All entries prior to %d have been nuked.", year))
					}
				} else {
					log.Printf("%s\n", color.YellowString("Nothing nuked."))
				}
			} else {
				log.Printf("%s\n", color.YellowString("Nothing nuked."))
			}
		} else {
			log.Printf("%s\n", color.YellowString("Nothing nuked."))
		}
	} else {
		cmd.Help()
	}
}

// Show the nuclear explosion on the screen.
// Concept taken from lazygit (https://github.com/jesseduffield/lazygit).
func showExplosion() {
	screen.Clear()
	width, height := screen.Size()

	max := 25
	for i := 0; i < max; i++ {
		screen.MoveTopLeft()
		fmt.Println(getExplodeImage(width, height, i, max))
		time.Sleep(time.Millisecond * 20)
	}
	screen.Clear()
}

// Render an explosion in the given bounds.
func getExplodeImage(width int, height int, frame int, max int) string {
	// Predefine the explosion symbols
	explosionChars := []rune{'*', '.', '@', '#', '&', '+', '%'}

	// Initialize a buffer to build our string.
	var buf bytes.Buffer

	// Initialize RNG seed.
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Calculate the center of explosion.
	centerX, centerY := width/2, height/2

	// Calculate the max radius (hypotenuse of the view).
	maxRadius := math.Hypot(float64(centerX), float64(centerY))

	// Calculate frame as a proportion of max, apply square root to create the non-linear effect.
	progress := math.Sqrt(float64(frame) / float64(max))

	// Calculate radius of explosion according to frame and max.
	radius := progress * maxRadius * 2

	// Introduce a new radius for the inner boundary of the explosion (the shockwave effect).
	var innerRadius float64
	if progress > 0.5 {
		innerRadius = (progress - 0.5) * 2 * maxRadius
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Calculate distance from center, scale x by 2 to compensate for character aspect ratio.
			distance := math.Hypot(float64(x-centerX), float64(y-centerY)*2)

			// If distance is less than radius and greater than innerRadius, draw explosion char.
			if distance <= radius && distance >= innerRadius {
				// Make placement random and less likely as explosion progresses.
				if random.Float64() > progress {
					// Pick a random explosion char.
					char := explosionChars[random.Intn(len(explosionChars))]
					buf.WriteRune(char)
				} else {
					buf.WriteRune(' ')
				}
			} else {
				// If not explosion, then it's empty space.
				buf.WriteRune(' ')
			}
		}
		// End of line.
		if y < height-1 {
			buf.WriteRune('\n')
		}
	}

	return buf.String()
}
