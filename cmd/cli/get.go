package cli

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/spf13/cobra"
)

type ReportTimeframe string

const (
	REPORT_TIMEFRAME_DAY   ReportTimeframe = "day"
	REPORT_TIMEFRAME_WEEK  ReportTimeframe = "week"
	REPORT_TIMEFRAME_MONTH ReportTimeframe = "month"
	REPORT_TIMEFRAME_YEAR  ReportTimeframe = "year"
)

var (
	weekReport      bool
	monthReport     string
	yearReport      string
	reportTimeframe *ReportTimeframe
	allReport       bool
	output          string
	filepath        string
)

var getCmd = &cobra.Command{
	Use:   "get [type]",
	Short: "Get a resource",
}

var getCompanyCmd = &cobra.Command{
	Use:     "company [name]",
	Short:   "Get a company",
	Aliases: []string{"companies"},
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 {
			company, err := CompanyRepository.GetByName(args[0])
			if err != nil {
				log.Fatalf("Unable to get company: %v", err)
			}
			fmt.Println(company.String())
		} else {
			companies, err := CompanyRepository.GetAll()
			if err != nil {
				log.Fatalf("Unable to get companies: %v", err)
			}
			for _, company := range companies {
				fmt.Println(company.String())
			}
		}
	},
}

var getSessionCmd = &cobra.Command{
	Use:   "session [date]",
	Short: "Get a work session",
	Long: `Get a work session. If no date is specified, the latest of current day is used.
    If a date is specified, the format must be YYYY-MM-DD.`,
	Example: `punch get session 
punch get session 2020-01-01
punch get session  01-01`,
	Aliases: []string{"sessions"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := getCompanyIfExists(currentCompanyName)
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

		reportTimeframe, err = getReportTimeframe()
		if err != nil {
			return err
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		slice, err := getRelevatSessions()
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}
		content := generateView(slice)
		fmt.Print(content)
	},
}

func getReportTimeframe() (*ReportTimeframe, error) {
	reportTimeframe := REPORT_TIMEFRAME_DAY
	if weekReport {
		reportTimeframe = REPORT_TIMEFRAME_WEEK
	} else if monthReport != "" {
		if err := validateMonth(monthReport); err != nil {
			return nil, err
		}
		reportTimeframe = REPORT_TIMEFRAME_MONTH
	} else if yearReport != "" {
		if err := validateYear(yearReport); err != nil {
			return nil, err
		}
		reportTimeframe = REPORT_TIMEFRAME_YEAR
	}
	return &reportTimeframe, nil
}

func validateMonth(month string) error {
	monthInt, err := strconv.Atoi(month)
	if err != nil {
		return fmt.Errorf("invalid month format")
	}
	if monthInt < 1 || monthInt > 12 {
		return fmt.Errorf("invalid month format")
	}
	return nil
}

func validateYear(year string) error {
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return fmt.Errorf("invalid year format")
	}
	currentYear := time.Now().Year()
	if yearInt < 1970 || yearInt > currentYear {
		return fmt.Errorf("invalid year format")
	}
	return nil
}

func getRelevatSessions() (*[]models.Session, error) {
	var slice []models.Session

	if *reportTimeframe == REPORT_TIMEFRAME_DAY {
		session, err := SessionRepository.GetLatestSessionOnSpecificDate(time.Now(), *currentCompany)
		if err != nil {
			return nil, err
		}
		slice = []models.Session{*session}
	} else {
		startDate := getStartDate()
		endDate := getEndDate(startDate)
		sessions, err := SessionRepository.GetAllSessions(*currentCompany)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}
		for _, session := range *sessions {
			if session.Start.After(startDate) && session.End.Before(endDate) {
				slice = append(slice, session)
			}
		}
	}
	return &slice, nil
}

func generateView(slice *[]models.Session) string {
	if len(*slice) == 0 {
		log.Fatalf("No available data")
	}
	var (
		buffer        *bytes.Buffer
		err           error
		content       string        = ""
		totalDuration time.Duration = 0
		totalEarnings float64       = 0
	)

	switch output {
	case "text":
		for _, session := range *slice {
			totalDuration += session.End.Sub(*session.Start)
			earnings, _ := session.Earnings()
			totalEarnings += earnings

			if verbose {
				content += session.Summary()
				content += "\n"
			}
		}

		prefix := getStartDate().Format("2006-01-02")
		if verbose {
			prefix = "Total\t"
		}
		content += fmt.Sprintf("%s\t%s\t%s\t%.2f %s\n",
			prefix,
			currentCompany.Name,
			totalDuration,
			totalEarnings,
			currentCompany.Currency,
		)
	case "csv":
		if verbose {
			buffer, err = models.SerializeSessionsToFullCSV(*slice)
		} else {
			buffer, err = models.SerializeSessionsToCSV(*slice)
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
	today := time.Now()
	year, _, _ := today.Date()

	switch *reportTimeframe {
	case REPORT_TIMEFRAME_WEEK:
		return today.AddDate(0, 0, -int(today.Weekday()))

	case REPORT_TIMEFRAME_YEAR:
		yr, err := parseYear(yearReport)
		if err == nil {
			return time.Date(yr, time.January, 1, 0, 0, 0, 0, today.Location())
		}

	case REPORT_TIMEFRAME_MONTH:
		mo, err := parseMonth(monthReport)
		if err == nil {
			return time.Date(year, mo, 1, 0, 0, 0, 0, today.Location())
		}
	}

	if allReport {
		return time.Time{}
	}

	return today.AddDate(0, 0, -today.Day()+1)
}

func getEndDate(startDate time.Time) time.Time {
	year, _, _ := startDate.Date()

	switch *reportTimeframe {
	case REPORT_TIMEFRAME_WEEK:
		return startDate.AddDate(0, 0, 6)

	case REPORT_TIMEFRAME_YEAR:
		yr, err := parseYear(yearReport)
		if err == nil {
			return time.Date(yr, time.December, 31, 0, 0, 0, 0, startDate.Location())
		}

	case REPORT_TIMEFRAME_MONTH:
		mo, err := parseMonth(monthReport)
		if err == nil {
			lastDay := lastDayOfMonth(year, mo)
			return time.Date(year, mo, lastDay, 0, 0, 0, 0, startDate.Location())
		}
	}

	if allReport {
		return time.Time{}
	}

	return startDate
}

func lastDayOfMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func getAmountOfTimeFilterFlags() int8 {
	flags := []bool{weekReport,
		monthReport != "",
		yearReport != "",
		allReport}
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
	currentYear, currentMonth, _ := time.Now().Date()
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getSessionCmd)
	getCmd.AddCommand(getCompanyCmd)
	getSessionCmd.Flags().StringVarP(&currentCompanyName, "company", "c", "", "Specify the company name")
	getSessionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	getSessionCmd.Flags().BoolVar(&weekReport, "week", false, "Get report for a specific week (format: YYYY-WW), leave empty for current week")
	getSessionCmd.Flags().StringVar(&monthReport, "month", "", "Get report for a specific month (format: YYYY-MM), leave empty for current month")
	getSessionCmd.Flags().StringVar(&yearReport, "year", "", "Get report for a specific year (format: YYYY), leave empty for current year")
	getSessionCmd.Flags().BoolVar(&allReport, "all", false, "Get all sessions")
	getSessionCmd.Flags().StringVarP(&output, "output", "o", "text", "Specify the output format")
	getSessionCmd.Flags().Lookup("month").NoOptDefVal = strconv.Itoa(int(currentMonth))
	getSessionCmd.Flags().Lookup("year").NoOptDefVal = strconv.Itoa(currentYear)
}
