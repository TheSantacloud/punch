package repositories

import (
	"errors"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound        = errors.New("Session record not found")
	SessionAlreadyExistsError = errors.New("Session already exists")
)

type RepoSession struct {
	ID          uint32 `gorm:"primaryKey;autoIncrement"`
	CompanyName string `gorm:"foreignKey:Name"`
	Start       *time.Time
	End         *time.Time
	Note        string
	Company     RepoCompany `gorm:"foreignKey:CompanyName;references:Name"`
}

type GORMSessionRepository struct {
	db *gorm.DB
}

func NewGORMSessionRepository(db *gorm.DB) *GORMSessionRepository {
	return &GORMSessionRepository{db}
}

func (repo *GORMSessionRepository) Insert(session *models.Session) error {
	repoSession := ToRepoSession(*session)

	var existingByID RepoSession
	idResult := repo.db.Where("id = ?", repoSession.ID).First(&existingByID)
	if idResult.Error == nil {
		return SessionAlreadyExistsError
	}

	var existingByDetails RepoSession
	detailResult := repo.db.Where("start = ? AND end = ? AND company_name = ?",
		repoSession.Start,
		repoSession.End,
		repoSession.CompanyName).First(&existingByDetails)

	if detailResult.Error == nil {
		return SessionAlreadyExistsError
	}

	return repo.db.Create(&repoSession).Error
}

func (repo *GORMSessionRepository) GetLatestSessionOnSpecificDateAllCompanies(date time.Time) (*[]models.Session, error) {
	var repoSessions []RepoSession
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := repo.db.Preload("Company").
		Where("start >= ? AND start < ?",
			startOfDay,
			endOfDay).
		Order("start DESC").
		First(&repoSessions).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	var sessions []models.Session
	for _, repoSession := range repoSessions {
		sessions = append(sessions, ToDomainSession(repoSession))
	}
	return &sessions, nil
}

func (repo *GORMSessionRepository) GetLatestSessionOnSpecificDate(date time.Time, company models.Company) (*models.Session, error) {
	var session RepoSession
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := repo.db.Preload("Company").
		Where("start >= ? AND start < ? AND company_name = ?",
			startOfDay,
			endOfDay,
			company.Name).
		Order("start DESC").
		First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	domainSession := ToDomainSession(session)
	return &domainSession, nil
}

func (repo *GORMSessionRepository) Update(session *models.Session) error {
	repoSession := ToRepoSession(*session)
	return repo.db.Save(&repoSession).Error
}

func (repo *GORMSessionRepository) GetAllSessions(company models.Company) (*[]models.Session, error) {
	var repoSessions []RepoSession
	err := repo.db.Preload("Company").
		Where("company_name = ?", company.Name).
		Order("start DESC").
		Find(&repoSessions).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	var sessions []models.Session
	for _, repoSession := range repoSessions {
		sessions = append(sessions, ToDomainSession(repoSession))
	}
	return &sessions, nil
}

func (repo *GORMSessionRepository) GetAllSessionsBetweenDates(company models.Company, start time.Time, end time.Time) (*[]models.Session, error) {
	var repoSessions []RepoSession
	err := repo.db.Preload("Company").
		Where("company_name = ? AND start >= ? AND end < ?",
			company.Name,
			start,
			end).
		Order("start DESC").
		Find(&repoSessions).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	var sessions []models.Session
	for _, repoSession := range repoSessions {
		sessions = append(sessions, ToDomainSession(repoSession))
	}
	return &sessions, nil
}

func (repo *GORMSessionRepository) GetAllSessionsAllCompanies() (*[]models.Session, error) {
	var repoSessions []RepoSession
	err := repo.db.Preload("Company").
		Order("start DESC").
		Find(&repoSessions).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	var sessions []models.Session
	for _, repoSession := range repoSessions {
		sessions = append(sessions, ToDomainSession(repoSession))
	}
	return &sessions, nil
}

func ToRepoSession(session models.Session) RepoSession {
	var companyName string
	if session.Company.Name != "" {
		companyName = session.Company.Name
	}

	return RepoSession{
		ID:          session.ID,
		CompanyName: companyName,
		Start:       session.Start,
		End:         session.End,
		Note:        session.Note,
		Company:     *ToRepoCompany(session.Company),
	}
}

func ToDomainSession(repoSession RepoSession) models.Session {
	return models.Session{
		ID:      repoSession.ID,
		Company: ToDomainCompany(repoSession.Company),
		Start:   repoSession.Start,
		End:     repoSession.End,
		Note:    repoSession.Note,
	}
}
