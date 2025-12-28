package regimen

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var bannerCmd = &cobra.Command{
	Use:   "banner",
	Short: "Display a themed ASCII art banner",
	Long: `Display a themed ASCII art banner.

The banner changes based on the current month.`,
	Run: runBanner,
}

var bannerMonth int

func init() {
	rootCmd.AddCommand(bannerCmd)
	bannerCmd.Flags().IntVarP(&bannerMonth, "month", "m", 0, "Override month (1-12)")
}

func runBanner(cmd *cobra.Command, args []string) {
	month := time.Now().Month()
	if bannerMonth >= 1 && bannerMonth <= 12 {
		month = time.Month(bannerMonth)
	}

	banner := getBanner(month)
	fmt.Println(banner)
}

func getBanner(month time.Month) string {
	switch month {
	case time.October:
		return bannerOctober()
	case time.December:
		return bannerDecember()
	default:
		return bannerDefault()
	}
}

func bannerOctober() string {
	return `
    .     .       .  .   . .   .   . .    +  .    .      .     .

 ██████  ██▓    ▓█████ ▓█████  ██▓███   ██▓    ▓█████   ██████   ██████
▒██    ▒ ▓██▒    ▓█   ▀ ▓█   ▀ ▓██░  ██▒▓██▒    ▓█   ▀ ▒██    ▒ ▒██    ▒
░ ▓██▄   ▒██░    ▒███   ▒███   ▓██░ ██▓▒▒██░    ▒███   ░ ▓██▄   ░ ▓██▄
  ▒   ██▒▒██░    ▒▓█  ▄ ▒▓█  ▄ ▒██▄█▓▒ ▒▒██░    ▒▓█  ▄   ▒   ██▒  ▒   ██▒
▒██████▒▒░██████▒░▒████▒░▒████▒▒██▒ ░  ░░██████▒░▒████▒▒██████▒▒▒██████▒▒
▒ ▒▓▒ ▒ ░░ ▒░▓  ░░░ ▒░ ░░░ ▒░ ░▒▓▒░ ░  ░░ ▒░▓  ░░░ ▒░ ░▒ ▒▓▒ ▒ ░▒ ▒▓▒ ▒ ░
░ ░▒  ░ ░░ ░ ▒  ░ ░ ░  ░ ░ ░  ░░▒ ░     ░ ░ ▒  ░ ░ ░  ░░ ░▒  ░ ░░ ░▒  ░ ░
░  ░  ░    ░ ░      ░      ░   ░░         ░ ░      ░   ░  ░  ░  ░  ░  ░
      ░      ░  ░   ░  ░   ░  ░             ░  ░   ░  ░      ░        ░

    .  :    .     .  :     .  :    .     .  :    .     .  :    .
`
}

func bannerDecember() string {
	return `
      *    *    	    *       *    		*        *      	 *    *

  ███████╗██╗     ███████╗███████╗██████╗ ██╗     ███████╗███████╗███████╗
  ██╔════╝██║     ██╔════╝██╔════╝██╔══██╗██║     ██╔════╝██╔════╝██╔════╝
  ███████╗██║     █████╗  █████╗  ██████╔╝██║     █████╗  ███████╗███████╗
  ╚════██║██║     ██╔══╝  ██╔══╝  ██╔═══╝ ██║     ██╔══╝  ╚════██║╚════██║
  ███████║███████╗███████╗███████╗██║     ███████╗███████╗███████║███████║
 ╚══════╝╚══════╝╚══════╝╚══════╝╚═╝     ╚══════╝╚══════╝╚══════╝╚══════╝

   *     *    * 	       * *     *    *	        * *   	  *    		*
`
}

func bannerDefault() string {
	return `
  ____  _     _____ _____ ____  _     _____ ____  ____
 / ___|| |   | ____| ____|  _ \| |   | ____/ ___||  ___/
 \___ \| |   |  _| |  _| | |_) | |   |  _| \___ \\___ \
  ___) | |___| |___| |___|  __/| |___| |___ ___) |__) |
 |____/|_____|_____|_____|_|   |_____|_____|____/____/
`
}
