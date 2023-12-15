package repositories

import (
	"github.com/dormunis/punch/pkg/models"
	"time"
)

type DayRepository interface {
	Insert(day *models.Day) error
	GetDayFromDateForAllCompanies(date time.Time) (*[]models.Day, error)
	GetDayFromDateForCompany(date time.Time, company models.Company) (*models.Day, error)
	Update(day *models.Day) error
	GetAllDaysForCompany(company models.Company) (*[]models.Day, error)
	GetAllDaysForAllCompanies() (*[]models.Day, error)
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
