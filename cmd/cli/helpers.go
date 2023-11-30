package cli

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

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
	if defaultCompany != "" && companyName == "" {
		companyName = defaultCompany
	}
	var err error
	company, err = timeTracker.GetCompany(companyName)
	if err != nil {
		return err
	}
	if company == nil {
		return fmt.Errorf("Company `%s` does not exist", name)
	}
	return nil
}
