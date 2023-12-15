package repositories

import (
	"github.com/dormunis/punch/pkg/models"
	"time"
)

type SessionRepository interface {
	Insert(session *models.Session) error
	Update(session *models.Session) error
	GetAllSessions(company models.Company) (*[]models.Session, error)
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
