package cli

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/viper"
)

func getParsedDateFromArgs(args []string) (time.Time, error) {
	var parsedDate time.Time
	var err error

	if len(args) == 0 {
		parsedDate = time.Now()
	} else {
		parsedDate, err = parseDate(args[0])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid date format")
		}
	}

	return parsedDate, nil
}

func parseDate(input string) (time.Time, error) {
	var layouts = []string{"2006-01-02", "2006-01", "2006"}
	var parsedDate time.Time
	var err error

	for _, layout := range layouts {
		parsedDate, err = time.Parse(layout, input)
		if err == nil {
			return parsedDate, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format")
}

func getParsedTimeFromArgs(args []string) (time.Time, error) {
	var parsedTime time.Time
	var err error

	if len(args) == 0 {
		parsedTime = time.Now()
	} else {
		parsedTime, err = parseTime(args[0])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid time format")
		}
	}

	return parsedTime, nil
}

func parseTime(input string) (time.Time, error) {
	var layouts = []string{"15:04:05", "15:04", "15"}
	var parsedTime time.Time
	var err error

	for _, layout := range layouts {
		parsedTime, err = time.Parse(layout, input)
		if err == nil {
			return parsedTime, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format")
}

func getCompanyIfExists(name string) error {
	defaultCompany := viper.GetString("settings.default_company")
	if defaultCompany != "" && currentCompanyName == "" {
		currentCompanyName = defaultCompany
	}
	var err error
	currentCompany, err = CompanyRepository.SafeGetByName(currentCompanyName)
	if err != nil {
		return err
	}
	if currentCompany == nil && currentCompanyName != defaultCompany {
		return fmt.Errorf("Company `%s` does not exist", currentCompanyName)
	} else if currentCompany == nil && currentCompanyName == defaultCompany {
		return fmt.Errorf("Set `%s` as default company, but it doesn't exists",
			currentCompanyName)
	}
	return nil
}

func parseYear(yearStr string) (int, error) {
	return strconv.Atoi(yearStr)
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
