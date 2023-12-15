package puncher

import (
	"errors"
	"fmt"
	"strings"

	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
)

var (
	ErrDayAlreadyStarted = errors.New("Day already started")
	ErrDayAlreadyEnded   = errors.New("Day already ended")
)

type Puncher struct {
	repo repositories.DayRepository
}

func NewPuncher(repo repositories.DayRepository) *Puncher {
	return &Puncher{
		repo: repo,
	}
}

func (p *Puncher) ToggleCheckInOut(company *models.Company, note string) (*models.Day, error) {
	today := time.Now()
	_, err := p.repo.GetDayFromDateForCompany(today, *company)
	switch err {
	case nil:
		return p.EndDay(*company, today, note)
	case repositories.ErrDayNotFound:
		return p.StartDay(*company, today, note)
	default:
		return nil, err
	}
}

func (p *Puncher) StartDay(company models.Company, timestamp time.Time, note string) (*models.Day, error) {
	if _, err := p.repo.GetDayFromDateForCompany(timestamp, company); err != repositories.ErrDayNotFound {
		return nil, ErrDayAlreadyStarted
	}
	day := models.Day{
		Company: company,
		Start:   &timestamp,
		Note:    note,
	}
	err := p.repo.Insert(&day)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, fmt.Errorf("Day already started; multiple starts not supported")
		}
		return nil, fmt.Errorf("Unable to insert day: %v", err)
	}
	return &day, nil
}

func (p *Puncher) EndDay(company models.Company, timestamp time.Time, note string) (*models.Day, error) {
	day, err := p.repo.GetDayFromDateForCompany(timestamp, company)
	switch err {
	case nil:
		if day.End != nil {
			return day, ErrDayAlreadyEnded
		}
	case repositories.ErrDayNotFound:
		return nil, ErrDayAlreadyEnded
	default:
		return nil, err
	}

	day.End = &timestamp

	if day.Note != "" {
		day.Note = day.Note + "; " + note
	} else {
		day.Note = note
	}

	err = p.repo.Update(day)
	if err != nil {
		return nil, err
	}

	return day, nil
}

// TODO: should this exist here?
func (p *Puncher) Sync(days *[]models.Day) error {
	// TODO: do this in bulk/async?
	for _, day := range *days {
		p.repo.Update(&day)
	}

	return nil
}
