package cli

import (
	"fmt"
	"strconv"
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

func getClientIfExists(name string) error {
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
