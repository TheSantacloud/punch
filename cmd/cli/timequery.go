package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dormunis/punch/pkg/models"
)

func ExtractTime(input string, client *models.Client) (time.Time, time.Time, error) {
	var parsedTime time.Time
	var err error

	if strings.HasPrefix(input, "-") {
		return extractFromTimeDelta(input)
	} else if strings.Contains(input, "HEAD~") {
		return extractFromCountDelta(input, client)
	}

	parsedTime, err = time.Parse("2006", input)
	if err == nil {
		startOfNextYear := parsedTime.AddDate(1, 0, 0).Truncate(24 * time.Hour)
		return parsedTime, startOfNextYear, nil
	}

	parsedTime, err = time.Parse("2006-01", input)
	if err == nil {
		startOfNextMonth := parsedTime.AddDate(0, 1, 0).Truncate(24 * time.Hour)
		return parsedTime, startOfNextMonth, nil
	}

	parsedTime, err = time.Parse("2006-01-02", input)
	if err == nil {
		startOfNextDay := parsedTime.Add(24 * time.Hour).Truncate(24 * time.Hour)
		return parsedTime, startOfNextDay, nil
	}

	return time.Time{}, time.Time{}, fmt.Errorf("invalid time format")
}

func extractFromTimeDelta(input string) (time.Time, time.Time, error) {
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
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time format")
	}

	delta, err := time.ParseDuration(input)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid time format")
	}

	parsedTime = time.Now().Add(delta)
	return parsedTime, time.Now(), nil
}

func extractFromCountDelta(input string, client *models.Client) (time.Time, time.Time, error) {
	count, err := strconv.Atoi(input[len("HEAD~"):])
	if err != nil || count < 1 {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid count format, must be ~<positive integer>")
	}

	sessions, err := SessionRepository.GetLastSessions(uint32(count), client)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	return *(*sessions)[len(*sessions)-1].Start, time.Now(), nil
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
