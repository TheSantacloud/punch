package puncher

import (
	"errors"
	"fmt"

	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
)

var (
	ErrSessionAlreadyStarted = errors.New("Session already started")
	ErrSessionAlreadyEnded   = errors.New("Session already ended")
	ErrInvalidSession        = errors.New("Invalid session")
)

type Puncher struct {
	repo repositories.SessionRepository
}

func NewPuncher(repo repositories.SessionRepository) *Puncher {
	return &Puncher{
		repo: repo,
	}
}

func (p *Puncher) ToggleCheckInOut(client *models.Client, note string) (*models.Session, error) {
	today := time.Now()
	session, err := p.repo.GetLatestSession()
	switch err {
	case nil:
		if session.Finished() {
			return p.StartSession(*client, today, note)
		} else {
			return p.EndSession(*session, today, note)
		}
	case repositories.ErrSessionNotFound:
		return p.StartSession(*client, today, note)
	default:
		return nil, err
	}
}

func (p *Puncher) StartSession(client models.Client, timestamp time.Time, note string) (*models.Session, error) {
	fetchedSession, err := p.repo.GetLatestSessionOnSpecificDate(timestamp, client)
	if err != repositories.ErrSessionNotFound && !fetchedSession.Finished() {
		return nil, ErrSessionAlreadyStarted
	}
	session := models.Session{
		Client: client,
		Start:  timestamp,
		Note:   note,
	}
	err = p.repo.Insert(&session, false)
	if err != nil {
		return nil, fmt.Errorf("Unable to insert session: %v", err)
	}
	return &session, nil
}

func (p *Puncher) EndSession(session models.Session, timestamp time.Time, note string) (*models.Session, error) {
	if session.Finished() {
		return nil, ErrSessionAlreadyEnded
	}

	if session.Start.After(timestamp) {
		return nil, ErrInvalidSession
	}

	session.End = timestamp

	lastyear := time.Date(time.Now().Year()-1, 1, 1, 0, 0, 0, 0, time.UTC).Year()
	timestampDaysSinceLastYear := int(timestamp.Sub(time.Date(lastyear, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24)
	startDaysSinceLastYear := int(session.Start.Sub(time.Date(lastyear, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24)
	if timestampDaysSinceLastYear > startDaysSinceLastYear {
		delta := timestampDaysSinceLastYear - startDaysSinceLastYear
		if note != "" {
			note += "; "
		}
		note += fmt.Sprintf("(+%d day)", int(delta))
	}

	if session.Note != "" {
		session.Note = session.Note + "; " + note
	} else {
		session.Note = note
	}

	err := p.repo.Update(&session, false)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
