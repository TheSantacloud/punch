/*
Copyright © 2023 Dor Munis <dormunis@gmail.com>
*/
package cli

import (
	"os"

	"fmt"

	"log"
	"time"

	"github.com/dormunis/consulting/cmd/daily"
	"github.com/dormunis/consulting/cmd/sheets"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "punch",
	Short: "Punchcard program for freelancers",
	Run: func(cmd *cobra.Command, args []string) {
		sheet, err := sheets.GetSheet()
		if err != nil {
			log.Fatalf("Unable to retrieve Sheets client: %v", err)
			os.Exit(1)
		}

		today, err := sheet.GetTodaysRow()
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}

		today.Update()
		sheet.UpdateRow(today)

		if today.End != "" {
			printSummary(today)
		}
	},
}

func printSummary(today *daily.Today) {
	start, _ := time.Parse("15:04:05", today.Start)
	end, _ := time.Parse("15:04:05", today.End)
	duration := end.Sub(start)
	moneyMade := fmt.Sprintf("+%.2f₪", duration.Hours()*sheets.PPH)
	fmt.Fprintf(os.Stdout, "[%s] Worked for %s (%s)\n", today.Date, duration, moneyMade)
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.consulting.yaml)")
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
