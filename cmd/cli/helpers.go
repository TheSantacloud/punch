package cli

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
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

func GetSessionsWithTimeframe(timeframe ReportTimeframe) []models.Session {
	var slice []models.Session

	startDate := getStartDate(timeframe)
	endDate := getEndDate(timeframe, startDate)
	sessions, err := SessionRepository.GetAllSessionsBetweenDates(startDate, endDate)
	if err != nil {
		return slice
	}
	for _, session := range *sessions {
		if currentClientName != "" && session.Client.Name != currentClientName {
			continue
		}
		if isWithinTimeframe(session.Start, startDate, endDate) &&
			(!session.Finished() || isWithinTimeframe(session.End, startDate, endDate)) {
			slice = append(slice, session)
		}
	}
	return slice
}

func GetRelativeSessionsFromArgs(args []string, clientName string) []models.Session {
	var slice []models.Session
	session, err := GetSessionByID(args[0])
	if err == nil {
		slice = append(slice, *session)
		return slice
	}

	startDate, endDate, err := ExtractParsedTimeFromArgs(args, clientName)
	if err != nil {
		return slice
	}
	sessions, err := SessionRepository.GetAllSessionsBetweenDates(startDate, endDate)
	if err != nil {
		return slice
	}
	for _, session := range *sessions {
		if currentClientName != "" && session.Client.Name != currentClientName {
			continue
		}
		if isWithinTimeframe(session.Start, startDate, endDate) &&
			(isWithinTimeframe(session.End, startDate, endDate) || !session.Finished()) {
			slice = append(slice, session)
		}
	}
	return slice
}

func GetSessionByID(id string) (*models.Session, error) {
	sessionId, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, err
	}

	session, err := SessionRepository.GetSessionByID(uint32(sessionId))
	if err != nil {
		return nil, err
	}
	return session, nil
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
			return prevSession.Start.After(nextSession.Start)
		} else {
			return prevSession.Start.Before(nextSession.Start)
		}
	})
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

func ExtractParsedTimeFromArgs(args []string, clientName string) (time.Time, time.Time, error) {
	var parsedTime time.Time
	var endTime time.Time
	var err error
	var client *models.Client

	if len(args) == 0 {
		parsedTime = time.Now()
	} else {
		if clientName == "" {
			client, err = ClientRepository.SafeGetByName(clientName)
			if err != nil {
				return time.Time{}, time.Time{}, err
			}
		}
		parsedTime, endTime, err = ExtractTime(args[0], client)
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid time format")
		}
	}

	return parsedTime, endTime, nil
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

func isWithinTimeframe(date time.Time, startDate time.Time, endDate time.Time) bool {
	return date.After(startDate) && date.Before(endDate)
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
