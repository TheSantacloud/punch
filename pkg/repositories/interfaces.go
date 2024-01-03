package repositories

import (
	"github.com/dormunis/punch/pkg/models"
	"time"
)

type SessionRepository interface {
	Insert(session *models.Session, dryRun bool) error
	Upsert(session *models.Session, dryRun bool) error
	Update(session *models.Session, dryRun bool) error
	Delete(session *models.Session, dryRun bool) error
	GetSessionByID(id uint32) (*models.Session, error)
	GetAllSessions(client models.Client) (*[]models.Session, error)
	GetAllSessionsBetweenDates(start time.Time, end time.Time) (*[]models.Session, error)
	GetAllSessionsAllClients() (*[]models.Session, error)
	GetLatestSession() (*models.Session, error)
	GetLatestSessionOnSpecificDate(date time.Time, client models.Client) (*models.Session, error)
	GetLatestSessionOnSpecificDateAllClients(date time.Time) (*[]models.Session, error)
	GetLastSessions(uint32, *models.Client) (*[]models.Session, error)
}

type ClientRepository interface {
	GetAll() ([]models.Client, error)
	Insert(client *models.Client) error
	Delete(client *models.Client) error
	GetByName(name string) (*models.Client, error)
	SafeGetByName(name string) (*models.Client, error)
	Rename(client *models.Client, newName string) error
	Update(client *models.Client) error
}
