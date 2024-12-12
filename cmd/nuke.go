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
	Short: "Nukes entries from the sqlite database",
	Long:  `As you continuously add completed entries, the database continues to grow unbounded.  The nuke command allows you to manage the size of your database.`,
	Run: func(cmd *cobra.Command, args []string) {
		runNuke(cmd, args)
	},
}

func init() {
	nukeCmd.Flags().BoolP(constants.ALL, constants.EMPTY, false, "Nuke ALL entries.  Use with extreme caution!!!")
	nukeCmd.Flags().BoolP(constants.PRIOR_YEARS, constants.EMPTY, false, "Nuke all entries prior to the current year's entries.")
	nukeCmd.Flags().BoolP(constants.DRY_RUN, constants.EMPTY, false, "Do not actually nuke anything, but show what potential would be nuked.")
	rootCmd.AddCommand(nukeCmd)
}

func runNuke(cmd *cobra.Command, _ []string) {
	all, _ := cmd.Flags().GetBool(constants.ALL)
	priorYears, _ := cmd.Flags().GetBool(constants.PRIOR_YEARS)
	dryRun, _ := cmd.Flags().GetBool(constants.DRY_RUN)

	if all {
		yesNo := yesNoPrompt("Are you sure you want to nuke ALL the entries from your database?")
		if yesNo {
			yesNo = yesNoPrompt("WARNING: Are you REALLY sure you want to nuke ALL the entries from your database?")
			if yesNo {
				yesNo = yesNoPrompt("LAST WARNING: Are you REALLY REALLY sure you want to nuke ALL the entries from your database?")
				if yesNo {
					// Yes was enter, so nuke ALL entries.
					db := database.New(viper.GetString(constants.DATABASE_FILE))
					var count = db.NukeAllEntries(dryRun)
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
					var count = db.NukePriorYearsEntries(dryRun, year)
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
