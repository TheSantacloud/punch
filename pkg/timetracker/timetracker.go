package timetracker

import (
	"fmt"
	"strings"

	"time"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/database"
	"github.com/dormunis/punch/pkg/sheetsync"
)

type TimeTracker struct {
	sheet *sheetsync.Sheet
	db    *database.Database

	syncOnEndDay bool
}

func NewTimeTracker(cfg *config.Config) (*TimeTracker, error) {
	db, err := database.NewDatabase(cfg.Database.Engine, cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("Unable to connect to database: %v", err)
	}

	sheet, err := sheetsync.GetSheet(*cfg.Sync.SpreadSheet)

	tt := TimeTracker{
		sheet:        sheet,
		db:           db,
		syncOnEndDay: cfg.Sync.SyncOnEndDay,
	}
	return &tt, nil
}

func (tt *TimeTracker) Sheet() {
	day, err := tt.GetDay(time.Now(), &database.Company{Name: "knostic"})
	err = tt.sheet.UpsertDay(*day)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func (tt *TimeTracker) ToggleCheckInOut(company *database.Company) error {
	today := time.Now()
	day, err := tt.db.GetDay(today, *company)
	if err != nil {
		return fmt.Errorf("Unable to get day: %v", err)
	}

	if day.Start == nil {
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

func (tt *TimeTracker) StartDay(company database.Company, timestamp time.Time) error {
	day := database.Day{
		Company: company,
		Start:   &timestamp,
	}
	err := tt.db.InsertNewDay(day)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return fmt.Errorf("Day already started; multiple starts not supported")
		}
		return fmt.Errorf("Unable to insert day: %v", err)
	}
	return nil
}

func (tt *TimeTracker) EndDay(company database.Company, timestamp time.Time) error {
	day, err := tt.db.GetDay(timestamp, company)
	if err != nil {
		return fmt.Errorf("Unable to get day: %v", err)
	}
	if day.End != nil {
		return fmt.Errorf("Day already ended")
	}
	day.End = &timestamp
	err = tt.db.UpdateDay(*day)
	if err != nil {
		return fmt.Errorf("Unable to update day: %v", err)
	}

	if tt.syncOnEndDay {
		err = tt.sheet.UpsertDay(*day)
		if err != nil {
			return fmt.Errorf("Unable to sync day: %v", err)
		}
	}
	return nil
}

func (tt *TimeTracker) SyncDay() error {
	// TODO: sync db to sheet
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
