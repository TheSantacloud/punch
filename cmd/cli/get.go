package cli

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dormunis/punch/pkg/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	weekReport  *bool
	monthReport *bool
	yearReport  *bool
	allReport   *bool
	output      string
	filepath    string
)

var getCmd = &cobra.Command{
	Use:   "get [type]",
	Short: "Get a resource",
}

var getCompanyCmd = &cobra.Command{
	Use:     "company [name]",
	Short:   "Get a company",
	Aliases: []string{"companies"},
	PreRun: func(cmd *cobra.Command, args []string) {
		getCompanyIfExists(companyName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if company != nil {
			company, err := timeTracker.GetCompany(company.Name)
			if err != nil {
				log.Fatalf("Unable to get company: %v", err)
			}
			fmt.Println(company.String())
		} else {
			companies, err := timeTracker.GetAllCompanies()
			if err != nil {
				log.Fatalf("Unable to get companies: %v", err)
			}
			for _, company := range *companies {
				fmt.Println(company.String())
			}
		}
	},
}

var getDayCmd = &cobra.Command{
	Use:   "day [date]",
	Short: "Get a work day",
	Long: `Get a work day. If no date is specified, the current day is used.
    If a date is specified, the format must be YYYY-MM-DD.`,
	Example: `punch get day
punch get day 2020-01-01
punch get day 01-01`,
	Aliases: []string{"days"},
	Args:    cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := getCompanyIfExists(companyName)
		if err != nil {
			// TODO: support all companies
			log.Fatalf("Report on all companies not supported yet")
		}
		err = preRunCheckOutput()
		if err != nil {
			return err
		}
		timeFlagCount := getAmountOfTimeFilterFlags()
		if timeFlagCount > 1 {
			return fmt.Errorf("only one of --week, --month, --year or --all can be set")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		slice := getRelevatDays()
		content := generateView(slice)
		fmt.Print(content)
	},
}

func getRelevatDays() *[]database.Day {
	var slice []database.Day
	timeFlagCount := getAmountOfTimeFilterFlags()

	if timeFlagCount == 0 {
		day, err := timeTracker.GetDay(time.Now(), company)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}
		slice = []database.Day{*day}
	} else {
		startDate := getStartDate()
		days, err := timeTracker.GetAllDays(company)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}
		for _, day := range *days {
			if day.Start.After(startDate) {
				slice = append(slice, day)
			}
		}
	}
	return &slice
}

func generateView(slice *[]database.Day) string {
	var (
		buffer        *bytes.Buffer
		err           error
		content       string        = ""
		totalDuration time.Duration = 0
		totalEarnings float64       = 0
	)

	switch output {
	case "text":
		for _, day := range *slice {
			totalDuration += day.End.Sub(*day.Start)
			earnings, _ := day.Earnings()
			totalEarnings += earnings

			if verbose {
				content += day.Summary()
				content += "\n"
			}
		}
		if !verbose {
			date := getStartDate().Format("2006-01-02")
			currency := viper.GetString("settings.currency")
			content += fmt.Sprintf("%s\t%s\t%s\t%s%.2f\n",
				date,
				company.Name,
				totalDuration,
				currency,
				totalEarnings)
		}
	case "csv":
		if verbose {
			buffer, err = database.SerializeDaysToFullCSV(*slice)
		} else {
			buffer, err = database.SerializeDaysToCSV(*slice)
		}
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}
		content = buffer.String()
	}
	return content
}

func getStartDate() time.Time {
	year, month, day := time.Now().Date()
	now := time.Date(year, month, day, 0, 0, 0, 0, time.Now().Location())
	if *weekReport {
		dayOfTheWeek := time.Now().Weekday()
		return now.AddDate(0, 0, -int(dayOfTheWeek))
	}
	if *yearReport {
		dayOfTheYear := time.Now().YearDay()
		return now.AddDate(0, 0, -int(dayOfTheYear)+1)
	}
	if *allReport {
		return time.Time{}
	}
	// default to month
	dayOfTheMonth := time.Now().Day()
	return now.AddDate(0, 0, -int(dayOfTheMonth)+1)
}

func getAmountOfTimeFilterFlags() int8 {
	flags := []bool{*weekReport, *monthReport, *yearReport, *allReport}
	setFlags := 0
	for _, flag := range flags {
		if flag {
			setFlags++
		}
	}

	return int8(setFlags)
}

func preRunCheckOutput() error {
	allowedOutputs := map[string]bool{"csv": true, "text": true}
	if _, ok := allowedOutputs[output]; !ok {
		return fmt.Errorf("invalid output format: %s, allowed formats are 'csv' and 'text'", output)
	}
	return nil

}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getCompanyCmd)
	getCmd.AddCommand(getDayCmd)
	getDayCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	getCompanyCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	getDayCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	// TODO: make month and year string flags with default values of current
	getDayCmd.Flags().BoolVar(weekReport, "week", false, "Get a specific week")
	getDayCmd.Flags().BoolVar(monthReport, "month", false, "Get a specific month")
	getDayCmd.Flags().BoolVar(yearReport, "year", false, "Get a specific year")
	getDayCmd.Flags().BoolVar(allReport, "all", false, "Get all days")
	getDayCmd.Flags().StringVarP(&output, "output", "o", "text", "Specify the output format")
}
