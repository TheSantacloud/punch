package repositories

import (
	"time"
)

func ToDateOnly(t time.Time) DateOnly {
	date := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return DateOnly{date}
}

func ToTimeOnly(t time.Time) TimeOnly {
	timePart := time.Date(0, 0, 0, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
	return TimeOnly{timePart}
}

func CombineDateAndTime(date time.Time, timePart time.Time) time.Time {
	return time.Date(
		date.Year(), date.Month(), date.Day(),
		timePart.Hour(), timePart.Minute(), timePart.Second(), timePart.Nanosecond(),
		date.Location(),
	)
}

type DateOnly struct {
	time.Time
}

type TimeOnly struct {
	time.Time
}
