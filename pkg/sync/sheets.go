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

func (s *SheetsSyncSource) Push(sessions *[]models.Session, approvedDiffs *[]models.Session) (PushSummary, error) {
	err := s.parseSheetIfNeeded()
	if err != nil {
		return PushSummary{}, err
	}

	mappedSessions := mapSessionsToRecords(sessions, s.cachedData)

	var sessionsToAdd []models.Session
	var recordsToUpdate []*sheets.Record
	var conflicts []models.Session

	for session, record := range mappedSessions {
		if record == nil {
			sessionsToAdd = append(sessionsToAdd, session)
		} else {
			if record.Session.ID == nil {
				record.Session = session
				recordsToUpdate = append(recordsToUpdate, record)
			} else if record.Session.Conflicts(session) {
				approved := false
				for _, diff := range *approvedDiffs {
					fmt.Printf("session.id: %d | diff.id: %d\n", *session.ID, *diff.ID)
					if *diff.ID == *session.ID {
						record.Session = session
						recordsToUpdate = append(recordsToUpdate, record)
						approved = true
						break
					}
				}
				if !approved {
					fmt.Printf("Conflict (ID: %v) between local and remote sessions\n", *session.ID)
					conflicts = append(conflicts, session)
				}
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

	if len(conflicts) > 0 {
		return PushSummary{}, fmt.Errorf("%d conflicts", len(conflicts))
	}

	// TODO: do this in bulk
	for _, session := range sessionsToAdd {
		err := s.Sheet.AddRow(session)
		if err != nil {
			return PushSummary{}, err
		}
	}

	// TODO: do this in bulk
	for _, record := range recordsToUpdate {
		err := s.Sheet.UpdateRow(*record)
		if err != nil {
			return PushSummary{}, err
		}
	}
	return PushSummary{
		Added:   len(sessionsToAdd),
		Updated: len(recordsToUpdate),
	}, nil
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
