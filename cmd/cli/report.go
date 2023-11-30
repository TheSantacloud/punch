package cli

import (
	"log"
	"time"

	"github.com/dormunis/punch/pkg/database"
	"github.com/spf13/cobra"
)

var (
	weekReport  bool
	monthReport bool
	yearReport  bool
	output      string
	filepath    string
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "report on work days",
	PreRun: func(cmd *cobra.Command, args []string) {
		err := getCompanyIfExists(companyName)
		if err != nil {
			// TODO: support all companies
			log.Fatalf("Report on all companies not supported yet")
		}
		flags := []bool{weekReport, monthReport, yearReport}
		setFlags := 0
		for _, flag := range flags {
			if flag {
				setFlags++
			}
		}
		if setFlags > 1 {
			log.Fatalf("only one of --week, --month, or --year can be set")
		}
		allowedOutputs := map[string]bool{"csv": true, "text": true}
		if _, ok := allowedOutputs[output]; !ok {
			log.Fatalf("invalid output format: %s, allowed formats are 'csv' and 'text'", output)
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		days, err := timeTracker.GetAllDays(company)
		if err != nil {
			log.Fatalf("%v", err)
		}
		startDate := getStartDate()
		var filtered *[]database.Day
		for _, day := range *days {
			if day.Start.After(startDate) {
				*filtered = append(*filtered, day)
			}
		}
		// TODO: keep going here to visualize it and save to file
	},
}

func getStartDate() time.Time {
	if weekReport {
		dayOfTheWeek := time.Now().Weekday()
		return time.Now().AddDate(0, 0, -int(dayOfTheWeek))
	}
	if yearReport {
		dayOfTheYear := time.Now().YearDay()
		return time.Now().AddDate(0, 0, -int(dayOfTheYear))
	}
	// default to month
	dayOfTheMonth := time.Now().Day()
	return time.Now().AddDate(0, 0, -int(dayOfTheMonth))
}

func init() {
	rootCmd.AddCommand(reportCmd)
	reportCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	reportCmd.Flags().BoolVar(&weekReport, "week", false, "Report on a specific week")
	reportCmd.Flags().BoolVar(&monthReport, "month", true, "Report on a specific month")
	reportCmd.Flags().BoolVar(&yearReport, "year", false, "Report on a specific year")
	reportCmd.Flags().StringVarP(&output, "output", "o", "csv", "Specify the output format")
	reportCmd.Flags().StringVarP(&filepath, "file", "f", "", "Specify the output file")
}
