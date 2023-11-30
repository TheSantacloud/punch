package database

import (
	"time"
)

type TimeOfDay string

const (
	BOD = TimeOfDay("Check In")
	EOD = TimeOfDay("Check Out")
)

type Period struct {
	TimeOfDay TimeOfDay
	Time      time.Time
}
