package sync

import (
	"fmt"

	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/dormunis/punch/pkg/sync/adapters/sheets"
)

type SheetsSyncSource struct {
	Sheet             *sheets.Sheet
	SessionRepository repositories.SessionRepository

	cachedData  *[]sheets.Record
	isDataFresh bool
}

func (s *SheetsSyncSource) Type() string {
	return "spreadsheet"
}

func (s *SheetsSyncSource) parseSheetIfNeeded() error {
	if s.isDataFresh {
		return nil
	}

	s.cachedData = &[]sheets.Record{}
	err := s.Sheet.ParseSheet(s.cachedData)
	if err != nil {
		return err
	}
	s.isDataFresh = true
	return nil
}

func (s *SheetsSyncSource) Pull() ([]models.Session, error) {
	err := s.parseSheetIfNeeded()
	if err != nil {
		return nil, err
	}

	var sessions []models.Session
	for _, record := range *s.cachedData {
		sessions = append(sessions, record.Session)
	}

	return sessions, nil
}

func (s *SheetsSyncSource) Push(sessions *[]models.Session) error {
	err := s.parseSheetIfNeeded()
	if err != nil {
		return err
	}

	mappedSessions := mapSessionsToRecords(sessions, s.cachedData)

	var sessionsToAdd []models.Session
	var recordsToUpdate []*sheets.Record

	for session, record := range mappedSessions {
		if record == nil {
			sessionsToAdd = append(sessionsToAdd, session)
		} else {
			if record.Session.ID == nil {
				record.Session = session
				recordsToUpdate = append(recordsToUpdate, record)
			} else if record.Session.Conflicts(session) {
				return fmt.Errorf("Session %v conflicts with %v",
					record.Session.String(), session.String())
			} else if *record.Session.ID != *session.ID {
				record.Session = session
				recordsToUpdate = append(recordsToUpdate, record)
			} else if record.Session.Equals(session) {
				continue
			} else {
				record.Session = session
				recordsToUpdate = append(recordsToUpdate, record)
			}
		}
	}

	if len(sessionsToAdd) == 0 {
		fmt.Println("No sessions to add")
	}

	// TODO: do this in bulk
	for _, session := range sessionsToAdd {
		err := s.Sheet.AddRow(session)
		if err != nil {
			return err
		}
		fmt.Printf("Added session %v\n", session.String())
	}

	if len(recordsToUpdate) == 0 {
		fmt.Println("No records to update")
	}
	// TODO: do this in bulk
	for _, record := range recordsToUpdate {
		err := s.Sheet.UpdateRow(*record)
		if err != nil {
			return err
		}
		fmt.Printf("Updated record %v\n", record.Session.String())
	}
	return nil
}

func mapSessionsToRecords(
	sessions *[]models.Session,
	records *[]sheets.Record) map[models.Session]*sheets.Record {
	mappedSessions := make(map[models.Session]*sheets.Record)
	for _, session := range *sessions {
		mappedSessions[session] = nil
		for _, record := range *records {
			if (record.Session.ID != nil && record.Session.ID == session.ID) ||
				record.Session.Similar(session) {
				mappedSessions[session] = &record
				break
			}
		}
	}
	return mappedSessions
}
