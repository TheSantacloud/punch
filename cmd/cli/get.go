package cli

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
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
	descendingOrder bool
)

var getCmd = &cobra.Command{
	Use:   "get [type]",
	Short: "Get a resource",
}

var getClientCmd = &cobra.Command{
	Use:     "client [name]",
	Short:   "Get a client",
	Aliases: []string{"clients"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			client, err := ClientRepository.GetByName(args[0])
			if err != nil {
				return fmt.Errorf("Unable to get client: %v", err)
			}
			fmt.Println(client.String())
		} else {
			clients, err := ClientRepository.GetAll()
			if err != nil {
				return fmt.Errorf("Unable to get clients: %v", err)
			}
			for _, client := range clients {
				fmt.Println(client.String())
			}
		}
		return nil
	},
}

var getSessionCmd = &cobra.Command{
	Use:   "session [date]",
	Short: "Get a work session",
	Long: `Get a work session. If no date is specified, the latest of current day is used.
    If a date is specified, the format must be YYYY-MM-DD.`,
	Example: `punch get session 
punch get session 2020-01-01
punch get session 01-01`,
	Aliases: []string{"sessions"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := preRunCheckOutput()
		if err != nil {
			return err
		}
		timeFlagCount := getAmountOfTimeFilterFlags()
		if timeFlagCount > 1 {
			return errors.New("only one of --week, --month, --year or --all can be set")
		}

		reportTimeframe, err = getTimeframe()
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		slice, err := getRelevatSessions()
		if err != nil {
			return err
		}

		sortSessions(slice, descendingOrder)

		content, err := generateView(slice)
		if err != nil {
			return err
		}
		fmt.Print(*content)
		return nil
	},
}

func getTimeframe() (*ReportTimeframe, error) {
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
		return errors.New("invalid month format")
	}
	if monthInt < 1 || monthInt > 12 {
		return errors.New("invalid month format")
	}
	return nil
}

func validateYear(year string) error {
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return errors.New("invalid year format")
	}
	currentYear := time.Now().Year()
	if yearInt < 1970 || yearInt > currentYear {
		return errors.New("invalid year format")
	}
	return nil
}

func getRelevatSessions() (*[]models.Session, error) {
	var slice []models.Session

	startDate := getStartDate()
	endDate := getEndDate(startDate)
	sessions, err := SessionRepository.GetAllSessionsBetweenDates(startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, session := range *sessions {
		if currentClientName != "" && session.Client.Name != currentClientName {
			continue
		}
		if session.Start.After(startDate) && session.End.Before(endDate) {
			slice = append(slice, session)
		}
	}
	return &slice, nil
}

func sortSessions(slice *[]models.Session, descending bool) {
	sort.SliceStable(*slice, func(i, j int) bool {
		prevSession := (*slice)[i]
		nextSession := (*slice)[j]
		if descending {
			return prevSession.Start.After(*nextSession.Start)
		} else {
			return prevSession.Start.Before(*nextSession.Start)
		}
	})
}

func generateView(slice *[]models.Session) (*string, error) {
	if len(*slice) == 0 {
		return nil, errors.New("No available data")
	}
	var (
		buffer        *bytes.Buffer
		err           error
		content       string        = ""
		totalDuration time.Duration = 0
		totalEarnings float64       = 0
		clientName    string        = ""
		currency      string        = ""
	)

	switch output {
	case "text":
		lastSessionDate := (*slice)[0].Start
		for _, session := range *slice {
			if session.End == nil {
				continue
			}
			totalDuration += session.End.Sub(*session.Start)
			earnings, _ := session.Earnings()
			totalEarnings += earnings
			// TODO: dont get just the latest client, figure out a better print here
			clientName = session.Client.Name
			currency = session.Client.Currency

			if session.Start.Day() > lastSessionDate.Day() {
				lastSessionDate = session.Start
			}

			if verbose {
				content += session.Summary()
				content += "\n"
			}
		}

		content += fmt.Sprintf("%s\t%s\t%s\t%.2f %s\n",
			lastSessionDate.Format("2006-01-02"),
			clientName,
			totalDuration,
			totalEarnings,
			currency,
		)
	case "csv":
		if verbose {
			buffer, err = models.SerializeSessionsToFullCSV(*slice)
		} else {
			buffer, err = models.SerializeSessionsToCSV(*slice)
		}
		if err != nil {
			return nil, err
		}
		content = buffer.String()
	}
	return &content, nil
}

func getStartDate() time.Time {
	today := time.Now()
	year, _, _ := today.Date()

	if allReport {
		return time.Time{}
	}

	switch *reportTimeframe {
	case REPORT_TIMEFRAME_DAY:
		return time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	case REPORT_TIMEFRAME_WEEK:
		daysToSubtract := int(today.Weekday())
		return time.Date(today.Year(), today.Month(), today.Day()-daysToSubtract, 0, 0, 0, 0, today.Location())
	case REPORT_TIMEFRAME_MONTH:
		mo, err := parseMonth(monthReport)
		if err == nil {
			return time.Date(year, mo, 1, 0, 0, 0, 0, today.Location())
		}
	case REPORT_TIMEFRAME_YEAR:
		yr, err := parseYear(yearReport)
		if err == nil {
			return time.Date(yr, time.January, 1, 0, 0, 0, 0, today.Location())
		}
	}

	// default to current day
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
			return time.Date(year, mo, lastDay, 0, 0, 0, 0, startDate.Location()).Add(24 * time.Hour)
		}
	}

	// default to now
	return time.Now()
}

func lastDayOfMonth(year int, month time.Month) int {
	if month > 12 {
		year += 1
	}
	month = month%12 + 1
	return time.Date(year, month, 0, 0, 0, 0, 0, time.UTC).Day()
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
	getCmd.AddCommand(getClientCmd)
	getSessionCmd.Flags().StringVarP(&currentClientName, "client", "c", "", "Specify the client name")
	getSessionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	getSessionCmd.Flags().BoolVar(&weekReport, "week", false, "Get report for a specific week (format: YYYY-WW), leave empty for current week")
	getSessionCmd.Flags().StringVar(&monthReport, "month", "", "Get report for a specific month (format: YYYY-MM), leave empty for current month")
	getSessionCmd.Flags().StringVar(&yearReport, "year", "", "Get report for a specific year (format: YYYY), leave empty for current year")
	getSessionCmd.Flags().BoolVar(&allReport, "all", false, "Get all sessions")
	getSessionCmd.Flags().BoolVar(&descendingOrder, "desc", false, "Sort sessions in descending order (defaults to ascending order)")
	getSessionCmd.Flags().StringVarP(&output, "output", "o", "text", "Specify the output format")
	getSessionCmd.Flags().Lookup("month").NoOptDefVal = strconv.Itoa(int(currentMonth))
	getSessionCmd.Flags().Lookup("year").NoOptDefVal = strconv.Itoa(currentYear)
}
