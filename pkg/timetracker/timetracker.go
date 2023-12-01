package timetracker

import (
	"fmt"
	"slices"
	"strings"

	"time"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/database"
	"github.com/dormunis/punch/pkg/sheetsync"

	"github.com/spf13/viper"
)

type TimeTracker struct {
	sheet *sheetsync.Sheet
	db    *database.Database
}

func NewTimeTracker(cfg *config.Config) (*TimeTracker, error) {
	db, err := database.NewDatabase(cfg.Database.Engine, cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %v", err)
	}

	sheet, err := sheetsync.GetSheet(*cfg.Sync.SpreadSheet)

	tt := TimeTracker{
		sheet: sheet,
		db:    db,
	}
	return &tt, nil
}

func (tt *TimeTracker) ToggleCheckInOut(company *database.Company) error {
	today := time.Now()
	_, err := tt.db.GetDay(today, *company)
	if err != nil {
		tt.StartDay(*company, today)
	} else {
		tt.EndDay(*company, today)
	}

	return nil
}

func (tt *TimeTracker) GetDay(datetime time.Time, company *database.Company) (*database.Day, error) {
	day, err := tt.db.GetDay(datetime, *company)
	if err != nil {
		return nil, fmt.Errorf("Unable to get day: %v", err)
	}
	return day, nil
}

func (tt *TimeTracker) GetAllDays(company *database.Company) (*[]database.Day, error) {
	days, err := tt.db.GetAllDays(*company)
	if err != nil {
		return nil, fmt.Errorf("Unable to get all days: %v", err)
	}
	return days, nil
}

func (tt *TimeTracker) StartDay(company database.Company, timestamp time.Time) (*database.Day, error) {
	day := database.Day{
		Company: company,
		Start:   &timestamp,
	}
	err := tt.db.InsertNewDay(day)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("Day already started; multiple starts not supported")
		}
		return nil, fmt.Errorf("Unable to insert day: %v", err)
	}
	if slices.Contains(viper.GetStringSlice("sync.sync_action"), "start") {
		slice := []database.Day{day}
		err = tt.Sync(&slice)
	}
	return &day, nil
}

func (tt *TimeTracker) EndDay(company database.Company, timestamp time.Time) (*database.Day, error) {
	day, err := tt.db.GetDay(timestamp, company)
	if err != nil {
		return nil, fmt.Errorf("Unable to get day: %v", err)
	}
	if day.End != nil {
		return nil, fmt.Errorf("Day already ended")
	}
	day.End = &timestamp
	err = tt.db.UpdateDay(*day)
	if err != nil {
		return nil, fmt.Errorf("Unable to update day: %v", err)
	}

	if slices.Contains(viper.GetStringSlice("sync.sync_action"), "end") {
		slice := []database.Day{*day}
		err = tt.Sync(&slice)
	}
	return day, nil
}

func (tt *TimeTracker) Sync(days *[]database.Day) error {
	var records []sheetsync.Record
	err := tt.sheet.ParseSheet(&records)
	if err != nil {
		return err
	}

	// TODO: do this in bulk/async?
	for _, day := range *days {
		tt.db.UpdateDay(day)
		err = tt.sheet.UpsertDay(day, &records)
		if err != nil {
			return fmt.Errorf("Unable to sync day: %v", err)
		}
	}

	return nil
}

func (tt *TimeTracker) GetCompany(name string) (*database.Company, error) {
	company, err := tt.db.GetCompany(name)
	if err != nil {
		return nil, fmt.Errorf("Unable to get company: %v", err)
	}
	return company, nil
}

func (tt *TimeTracker) GetAllCompanies() (*[]database.Company, error) {
	companies, err := tt.db.GetAllCompanies()
	if err != nil {
		return nil, fmt.Errorf("Unable to get all companies: %v", err)
	}
	return companies, nil
}

func (tt *TimeTracker) AddCompany(name string, pph int32) error {
	company := database.Company{
		Name: name,
		PPH:  pph,
	}
	err := tt.db.InsertCompany(company)
	if err != nil {
		return fmt.Errorf("Unable to insert company: %v", err)
	}
	return nil
}
