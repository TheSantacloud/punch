package repositories

import (
	"errors"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"gorm.io/gorm"
)

var ErrDayNotFound = errors.New("record not found")

type RepoDay struct {
	ID          uint32 `gorm:"primaryKey;autoIncrement"`
	CompanyName string `gorm:"foreignKey:Name"`
	Start       *time.Time
	End         *time.Time
	Note        string
	Company     RepoCompany `gorm:"foreignKey:CompanyName;references:Name"`
}

type DayRepository interface {
	Insert(day *models.Day) error
	GetDayFromDateForAllCompanies(date time.Time) (*[]models.Day, error)
	GetDayFromDateForCompany(date time.Time, company models.Company) (*models.Day, error)
	Update(day *models.Day) error
	GetAllDaysForCompany(company models.Company) (*[]models.Day, error)
	GetAllDaysForAllCompanies() (*[]models.Day, error)
}

type GORMDayRepository struct {
	db *gorm.DB
}

func NewGORMDayRepository(db *gorm.DB) *GORMDayRepository {
	return &GORMDayRepository{db}
}

func (repo *GORMDayRepository) Insert(day *models.Day) error {
	repoDay := ToRepoDay(*day)
	return repo.db.Create(&repoDay).Error
}

func (repo *GORMDayRepository) GetDayFromDateForAllCompanies(date time.Time) (*[]models.Day, error) {
	var repoDays []RepoDay
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := repo.db.Preload("Company").Where("start >= ? AND start < ?", startOfDay, endOfDay).First(&repoDays).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrDayNotFound
		}
		return nil, err
	}
	var days []models.Day
	for _, repoDay := range repoDays {
		days = append(days, ToDomainDay(repoDay))
	}
	return &days, nil
}

func (repo *GORMDayRepository) GetDayFromDateForCompany(date time.Time, company models.Company) (*models.Day, error) {
	var day RepoDay
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := repo.db.Preload("Company").Where("start >= ? AND start < ? AND company_name = ?", startOfDay, endOfDay, company.Name).First(&day).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrDayNotFound
		}
		return nil, err
	}
	domainDay := ToDomainDay(day)
	return &domainDay, nil
}

func (repo *GORMDayRepository) Update(day *models.Day) error {
	repoDay := ToRepoDay(*day)
	return repo.db.Save(&repoDay).Error
}

func (repo *GORMDayRepository) GetAllDaysForCompany(company models.Company) (*[]models.Day, error) {
	var repoDays []RepoDay
	err := repo.db.Preload("Company").Where("company_name = ?", company.Name).Find(&repoDays).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrDayNotFound
		}
		return nil, err
	}
	var days []models.Day
	for _, repoDay := range repoDays {
		days = append(days, ToDomainDay(repoDay))
	}
	return &days, nil
}

func (repo *GORMDayRepository) GetAllDaysForAllCompanies() (*[]models.Day, error) {
	var repoDays []RepoDay
	err := repo.db.Preload("Company").Find(&repoDays).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrDayNotFound
		}
		return nil, err
	}
	var days []models.Day
	for _, repoDay := range repoDays {
		days = append(days, ToDomainDay(repoDay))
	}
	return &days, nil
}

func ToRepoDay(day models.Day) RepoDay {
	var companyName string
	if day.Company.Name != "" {
		companyName = day.Company.Name
	}

	return RepoDay{
		ID:          day.ID,
		CompanyName: companyName,
		Start:       day.Start,
		End:         day.End,
		Note:        day.Note,
		Company:     *ToRepoCompany(day.Company),
	}
}

func ToDomainDay(repoDay RepoDay) models.Day {
	return models.Day{
		ID:      repoDay.ID,
		Company: ToDomainCompany(repoDay.Company),
		Start:   repoDay.Start,
		End:     repoDay.End,
		Note:    repoDay.Note,
	}
}
