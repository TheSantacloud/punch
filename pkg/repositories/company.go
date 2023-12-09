package repositories

import (
	"errors"

	"github.com/dormunis/punch/pkg/models"
	"gorm.io/gorm"
)

var ErrCompanyNotFound = errors.New("record not found")

type RepoCompany struct {
	Name     string `gorm:"primaryKey"`
	PPH      uint16
	Currency string
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

type GORMCompanyRepository struct {
	db *gorm.DB
}

func NewGORMCompanyRepository(db *gorm.DB) *GORMCompanyRepository {
	return &GORMCompanyRepository{db}
}

func (repo *GORMCompanyRepository) GetAll() ([]models.Company, error) {
	var repoCompanies []RepoCompany
	err := repo.db.Find(&repoCompanies).Error
	if err != nil {
		return nil, err
	}
	var companies []models.Company
	for _, repoCompany := range repoCompanies {
		companies = append(companies, ToDomainCompany(repoCompany))
	}
	return companies, nil
}

func (repo *GORMCompanyRepository) Insert(company *models.Company) error {
	repoCompany := ToRepoCompany(*company)
	return repo.db.FirstOrCreate(repoCompany, RepoCompany{Name: repoCompany.Name}).Error
}

func (repo *GORMCompanyRepository) Delete(company *models.Company) error {
	repoCompany := ToRepoCompany(*company)
	return repo.db.Delete(repoCompany).Error
}

func (repo *GORMCompanyRepository) GetByName(name string) (*models.Company, error) {
	var repoCompany RepoCompany
	err := repo.db.Where("name = ?", name).First(&repoCompany).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrCompanyNotFound
		}
		return nil, err
	}
	company := ToDomainCompany(repoCompany)
	return &company, nil
}

func (repo *GORMCompanyRepository) SafeGetByName(name string) (*models.Company, error) {
	var company RepoCompany
	err := repo.db.Where("name = ?", name).First(&company).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	domainCompany := ToDomainCompany(company)
	return &domainCompany, nil
}

func (repo *GORMCompanyRepository) Rename(company *models.Company, newName string) error {
	company.Name = newName
	repoCompany := ToRepoCompany(*company)
	return repo.db.Save(repoCompany).Error
}

func (repo *GORMCompanyRepository) Update(company *models.Company) error {
	repoCompany := ToRepoCompany(*company)
	return repo.db.Save(repoCompany).Error
}

func ToRepoCompany(company models.Company) *RepoCompany {
	return &RepoCompany{
		Name:     company.Name,
		PPH:      company.PPH,
		Currency: company.Currency,
	}
}

func ToDomainCompany(company RepoCompany) models.Company {
	return models.Company{
		Name:     company.Name,
		PPH:      company.PPH,
		Currency: company.Currency,
	}
}
