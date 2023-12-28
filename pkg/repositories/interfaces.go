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
	GetAllSessions(company models.Company) (*[]models.Session, error)
	GetAllSessionsBetweenDates(company models.Company, start time.Time, end time.Time) (*[]models.Session, error)
	GetAllSessionsAllCompanies() (*[]models.Session, error)
	GetLatestSessionOnSpecificDate(date time.Time, company models.Company) (*models.Session, error)
	GetLatestSessionOnSpecificDateAllCompanies(date time.Time) (*[]models.Session, error)
}

type CompanyRepository interface {
	GetAll() ([]models.Company, error)
	Insert(company *models.Company) error
	Delete(company *models.Company) error
	GetByName(name string) (*models.Company, error)
	SafeGetByName(name string) (*models.Company, error)
	Rename(company *models.Company, newName string) error
	Update(company *models.Company) error
}
