package sync

import (
	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/dormunis/punch/pkg/sync/adapters/sheets"
)

type SheetsSyncSource struct {
	Sheet             *sheets.Sheet
	SessionRepository repositories.SessionRepository
}

func (s *SheetsSyncSource) Type() string {
	return "spreadsheet"
}

func (s *SheetsSyncSource) Pull() (*[]models.Session, error) {
	var records []sheets.Record
	err := s.Sheet.ParseSheet(&records)
	if err != nil {
		return nil, err
	}

	var conflicts []models.Session
	for _, record := range records {
		err = s.SessionRepository.Insert(&record.Session)
		if err == nil {
			continue
		} else if err != nil && err != repositories.SessionAlreadyExistsError {
			conflicts = append(conflicts, record.Session)
		}
	}

	return &conflicts, nil
}

func (s *SheetsSyncSource) Push() (*[]models.Session, error) {
	localSessions, err := s.SessionRepository.GetAllSessionsAllCompanies()
	if err != nil {
		return nil, err
	}

	var sheetRecords []sheets.Record
	err = s.Sheet.ParseSheet(&sheetRecords)
	if err != nil {
		return nil, err
	}

	var conflicts []models.Session
	for _, localSession := range *localSessions {
		if !sessionExistsInSheet(localSession, sheetRecords) {
			err := s.Sheet.AddRow(localSession)
			if err != nil {
				return nil, err
			}
		} else if sessionConflicts(localSession, sheetRecords) {
			conflicts = append(conflicts, localSession)
		}
	}

	return &conflicts, nil
}

func sessionExistsInSheet(session models.Session, records []sheets.Record) bool {
	for _, record := range records {
		if record.Matches(session) {
			return true
		}
	}
	return false
}

func sessionConflicts(session models.Session, records []sheets.Record) bool {
	for _, record := range records {
		if record.Session.Start.Format("02/01/2006") == session.Start.Format("02/01/2006") &&
			(record.Session.Start != session.Start || record.Session.End != session.End ||
				record.Session.Company.Name != session.Company.Name) {
			return true
		}
	}
	return false
}
