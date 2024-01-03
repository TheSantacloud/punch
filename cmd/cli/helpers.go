package cli

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/spf13/viper"
)

type ReportTimeframe string

const (
	REPORT_TIMEFRAME_DAY   ReportTimeframe = "day"
	REPORT_TIMEFRAME_WEEK  ReportTimeframe = "week"
	REPORT_TIMEFRAME_MONTH ReportTimeframe = "month"
	REPORT_TIMEFRAME_YEAR  ReportTimeframe = "year"
)

var (
	clientName      string
	dayReport       bool
	weekReport      bool
	monthReport     string
	yearReport      string
	reportTimeframe *ReportTimeframe
	allReport       bool
)

func GetSessionsWithTimeframe(timeframe ReportTimeframe) (*[]models.Session, error) {
	var slice []models.Session

	startDate := getStartDate(timeframe)
	endDate := getEndDate(timeframe, startDate)
	sessions, err := SessionRepository.GetAllSessionsBetweenDates(startDate, endDate)
	if err != nil {
		return nil, err
	}
	for _, session := range *sessions {
		if currentClientName != "" && session.Client.Name != currentClientName {
			continue
		}
		if session.Start.After(startDate) &&
			(session.End == nil ||
				(session.End != nil && session.End.Before(endDate))) {
			slice = append(slice, session)
		}
	}
	return &slice, nil
}

func GetRelativeSessionsFromArgs(args []string, clientName string) (*[]models.Session, error) {
	startDate, err := ExtractParsedTimeFromArgs(args, clientName)
	if err != nil {
		return nil, err
	}
	return SessionRepository.GetAllSessionsBetweenDates(startDate, time.Now())
}

func FilterSessionsByClient(sessions *[]models.Session, clientName string) *[]models.Session {
	if clientName == "" {
		return sessions
	}
	var filteredSessions []models.Session
	for _, session := range *sessions {
		if session.Client.Name == clientName {
			filteredSessions = append(filteredSessions, session)
		}
	}
	return &filteredSessions
}

func SortSessions(slice *[]models.Session, descending bool) {
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

func getStartDate(timeframe ReportTimeframe) time.Time {
	today := time.Now()
	year, _, _ := today.Date()

	if allReport {
		return time.Time{}
	}

	switch timeframe {
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

func getEndDate(timeframe ReportTimeframe, startDate time.Time) time.Time {
	year, _, _ := startDate.Date()

	switch timeframe {
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

func parseMonth(monthStr string) (time.Month, error) {
	monthInt, err := strconv.Atoi(monthStr)
	if err != nil {
		return 0, err
	}
	if monthInt < 1 || monthInt > 12 {
		return 0, fmt.Errorf("invalid month")
	}
	return time.Month(monthInt), nil
}

func parseYear(yearStr string) (int, error) {
	return strconv.Atoi(yearStr)
}

func lastDayOfMonth(year int, month time.Month) int {
	if month > 12 {
		year += 1
	}
	month = month%12 + 1
	return time.Date(year, month, 0, 0, 0, 0, 0, time.UTC).Day()
}

func ExtractTimeframeFromFlags() (*ReportTimeframe, error) {
	timeFlagCount := getAmountOfTimeFilterFlags()
	if timeFlagCount > 1 {
		return nil, errors.New("only one of --day, --week, --month, --year or --all can be set")
	}

	reportTimeframe := REPORT_TIMEFRAME_DAY
	if dayReport {
		reportTimeframe = REPORT_TIMEFRAME_DAY
	} else if weekReport {
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

func getAmountOfTimeFilterFlags() int8 {
	flags := []bool{
		dayReport,
		weekReport,
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

func ExtractParsedTimeFromArgs(args []string, clientName string) (time.Time, error) {
	var parsedTime time.Time
	var err error
	var client *models.Client

	if len(args) == 0 {
		parsedTime = time.Now()
	} else {
		if clientName == "" {
			client, err = ClientRepository.SafeGetByName(clientName)
			if err != nil {
				return time.Time{}, err
			}
		}
		parsedTime, err = extractTime(args[0], client)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time format")
		}
	}

	return parsedTime, nil
}

func extractTime(input string, client *models.Client) (time.Time, error) {
	var layouts = []string{"15:04:05", "15:04", "15"}
	var parsedTime time.Time
	var err error

	if strings.HasPrefix(input, "-") {
		return extractFromTimeDelta(input)
	} else if strings.Contains(input, "HEAD~") {
		return extractFromCountDelta(input, client)
	}

	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, input)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format")
}

func extractFromTimeDelta(input string) (time.Time, error) {
	var parsedTime time.Time
	var err error

	suffix := input[len(input)-1:]
	switch suffix {
	case "d":
		input, err = parseDurationMoreThanHour(input, 24)
	case "w":
		input, err = parseDurationMoreThanHour(input, 24*7)
	case "M":
		input, err = parseDurationMoreThanHour(input, 24*30)
	case "y":
		input, err = parseDurationMoreThanHour(input, 24*365)
	}
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format")
	}

	delta, err := time.ParseDuration(input)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid time format")
	}

	parsedTime = time.Now().Add(delta)
	return parsedTime, nil
}

func extractFromCountDelta(input string, client *models.Client) (time.Time, error) {
	count, err := strconv.Atoi(input[len("HEAD~"):])
	if err != nil || count < 1 {
		return time.Time{}, fmt.Errorf("invalid count format, must be ~<positive integer>")
	}

	sessions, err := SessionRepository.GetLastSessions(uint32(count), client)
	if err != nil {
		return time.Time{}, err
	}

	return *(*sessions)[len(*sessions)-1].Start, nil
}

func parseDurationMoreThanHour(input string, multiplier int) (string, error) {
	input = input[1 : len(input)-1]
	number, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return "", fmt.Errorf("invalid time format")
	}
	delta := time.Duration(number*float64(multiplier)) * time.Hour
	return fmt.Sprintf("-%s", delta.String()), nil
}

func GetClientIfExists(name string) error {
	defaultClient := viper.GetString("settings.default_client")
	if defaultClient != "" && currentClientName == "" {
		currentClientName = defaultClient
	}
	var err error
	currentClient, err = ClientRepository.SafeGetByName(currentClientName)
	if err != nil {
		return err
	}
	if currentClient == nil && currentClientName != defaultClient {
		return fmt.Errorf("Client `%s` does not exist", currentClientName)
	} else if currentClient == nil && currentClientName == defaultClient {
		return fmt.Errorf("Set `%s` as default client, but it doesn't exists",
			currentClientName)
	}
	return nil
}
