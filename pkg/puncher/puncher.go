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
	session, err := p.repo.GetLatestSessionOnSpecificDate(today, *client)
	switch err {
	case nil:
		if session.End != nil {
			return p.StartSession(*client, today, note)
		} else {
			return p.EndSession(*client, today, note)
		}
	case repositories.ErrSessionNotFound:
		return p.StartSession(*client, today, note)
	default:
		return nil, err
	}
}

func (p *Puncher) StartSession(client models.Client, timestamp time.Time, note string) (*models.Session, error) {
	fetchedSession, err := p.repo.GetLatestSessionOnSpecificDate(timestamp, client)
	if err != repositories.ErrSessionNotFound && fetchedSession.End == nil {
		return nil, ErrSessionAlreadyStarted
	}
	session := models.Session{
		Client: client,
		Start:  &timestamp,
		Note:   note,
	}
	err = p.repo.Insert(&session, false)
	if err != nil {
		return nil, fmt.Errorf("Unable to insert session: %v", err)
	}
	return &session, nil
}

func (p *Puncher) EndSession(client models.Client, timestamp time.Time, note string) (*models.Session, error) {
	session, err := p.repo.GetLatestSessionOnSpecificDate(timestamp, client)
	switch err {
	case nil:
		if session.End != nil {
			return session, ErrSessionAlreadyEnded
		}
	case repositories.ErrSessionNotFound:
		return nil, ErrSessionAlreadyEnded
	default:
		return nil, err
	}

	session.End = &timestamp

	if session.Note != "" {
		session.Note = session.Note + "; " + note
	} else {
		session.Note = note
	}

	err = p.repo.Update(session, false)
	if err != nil {
		return nil, err
	}

	return session, nil
}
