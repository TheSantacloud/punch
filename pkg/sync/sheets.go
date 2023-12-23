package sync

import (
	"errors"
	"fmt"

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

func (s *SheetsSyncSource) Pull() (map[models.Session]error, error) {
	var records []sheets.Record
	// TODO: this is called twice (pull and push)
	err := s.Sheet.ParseSheet(&records)
	if err != nil {
		return nil, err
	}

	var sessionsToInsert []models.Session

	var conflicts map[models.Session]error
	conflicts = make(map[models.Session]error)

	for _, record := range records {
		err = s.SessionRepository.Insert(&record.Session, true)
		if err != nil &&
			!errors.Is(err, repositories.ConflictingIdsError) {
			conflicts[record.Session] = err
		} else if err == nil {
			sessionsToInsert = append(sessionsToInsert, record.Session)
		}
	}

	if len(sessionsToInsert) == 0 {
		fmt.Println("No sessions to insert")
	} else {
		fmt.Printf("Inserting %d sessions\n", len(sessionsToInsert))
	}

	if len(conflicts) == 0 {
		for _, session := range sessionsToInsert {
			err = s.SessionRepository.Insert(&session, false)
			if err != nil {
				return nil, err
			}
		}
	}

	return conflicts, nil
}

func (s *SheetsSyncSource) Push() (map[models.Session]error, error) {
	localSessions, err := s.SessionRepository.GetAllSessionsAllCompanies()
	if err != nil {
		return nil, err
	}

	var sheetRecords []sheets.Record
	err = s.Sheet.ParseSheet(&sheetRecords)
	if err != nil {
		return nil, err
	}

	mappedSessions := mapSessionsToRecords(localSessions, &sheetRecords)
	var conflicts map[models.Session]error
	conflicts = make(map[models.Session]error)

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
				conflicts[session] = fmt.Errorf("Session %v conflicts with %v", session, record.Session)
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

	if len(conflicts) == 0 {
		if len(sessionsToAdd) == 0 {
			fmt.Println("No sessions to add")
		} else {
			fmt.Printf("Adding %d sessions\n", len(sessionsToAdd))
		}

		// TODO: do this in bulk
		for _, session := range sessionsToAdd {
			// TODO: verbose print all records to add
			err := s.Sheet.AddRow(session)
			if err != nil {
				return nil, err
			}
		}

		if len(recordsToUpdate) == 0 {
			fmt.Println("No records to update")
		} else {
			fmt.Printf("Updating %d records\n", len(recordsToUpdate))
		}
		// TODO: do this in bulk
		for _, record := range recordsToUpdate {
			// TODO: verbose print all records to update
			err := s.Sheet.UpdateRow(*record)
			if err != nil {
				return nil, err
			}
		}
	}

	return conflicts, nil
}

func mapSessionsToRecords(sessions *[]models.Session, records *[]sheets.Record) map[models.Session]*sheets.Record {
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
