package repositories

import (
	"errors"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound  = errors.New("Session record not found")
	ConflictingIdsError = errors.New("Session already exists with a different ID")
	InfoConflictError   = errors.New("Session exists with different info")
)

type RepoSession struct {
	ID          *uint32 `gorm:"primaryKey;autoIncrement"`
	CompanyName string  `gorm:"foreignKey:Name"`
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

func (repo *GORMSessionRepository) Insert(session *models.Session, dryRun bool) error {
	repoSession := ToRepoSession(*session)

	if session.ID != nil {
		sessionByID, err := repo.GetSessionByID(*session.ID)
		if err == nil {
			if session.ID != nil &&
				sessionByID.ID != nil &&
				sessionByID.Conflicts(*session) {
				return InfoConflictError
			}
		}
	}

	var existingByDetails RepoSession
	detailResult := repo.db.Where("start = ? AND company_name = ?",
		repoSession.Start,
		repoSession.CompanyName).First(&existingByDetails)

	if detailResult.Error == nil && (session.ID == nil || existingByDetails.ID != session.ID) {
		return ConflictingIdsError
	}

	if detailResult.Error == nil {
		sessionByDetails := ToDomainSession(existingByDetails)
		if sessionByDetails.Conflicts(*session) {
			return InfoConflictError
		}
	}

	if dryRun {
		return repo.db.Session(&gorm.Session{DryRun: true}).Create(&repoSession).Error
	}
	return repo.db.Create(&repoSession).Error
}

func (repo *GORMSessionRepository) Upsert(session *models.Session, dryRun bool) error {
	repoSession := ToRepoSession(*session)
	if dryRun {
		return repo.db.Session(&gorm.Session{DryRun: true}).Save(&repoSession).Error
	}

	if session.ID == nil {
		var existingByDetails RepoSession
		detailResult := repo.db.Where("start = ? AND company_name = ?",
			repoSession.Start,
			repoSession.CompanyName).First(&existingByDetails)

		if detailResult.Error != nil {
			session.ID = existingByDetails.ID
		}
	}

	return repo.db.Save(&repoSession).Error
}

func (repo *GORMSessionRepository) GetSessionByID(id uint32) (*models.Session, error) {
	var repoSession RepoSession
	err := repo.db.Preload("Company").Where("id = ?", id).First(&repoSession).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	domainSession := ToDomainSession(repoSession)
	return &domainSession, nil
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

func (repo *GORMSessionRepository) Update(session *models.Session, dryRun bool) error {
	repoSession := ToRepoSession(*session)
	if dryRun {
		return repo.db.Session(&gorm.Session{DryRun: true}).Save(&repoSession).Error
	}
	return repo.db.Save(&repoSession).Error
}

func (repo *GORMSessionRepository) Delete(session *models.Session, dryRun bool) error {
	repoSession := ToRepoSession(*session)
	if dryRun {
		return repo.db.Session(&gorm.Session{DryRun: true}).Delete(&repoSession).Error
	}
	return repo.db.Delete(&repoSession).Error
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

func (repo *GORMSessionRepository) GetAllSessionsBetweenDates(start time.Time, end time.Time) (*[]models.Session, error) {
	var repoSessions []RepoSession
	err := repo.db.Preload("Company").
		Where("start >= ?", start).
		Where("end < ? OR (end IS NULL OR end = '')", end).
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

	startTime := session.Start.Truncate(time.Second)
	endTime := session.End
	if endTime != nil {
		*endTime = endTime.Truncate(time.Second)
	}

	return RepoSession{
		ID:          session.ID,
		CompanyName: companyName,
		Start:       &startTime,
		End:         endTime,
		Note:        session.Note,
		Company:     *ToRepoCompany(session.Company),
	}
}

func ToDomainSession(repoSession RepoSession) models.Session {
	startTime := repoSession.Start
	var endTime *time.Time
	if repoSession.End != nil {
		endTime = new(time.Time)
		*endTime = *repoSession.End
	}
	return models.Session{
		ID:      repoSession.ID,
		Company: ToDomainCompany(repoSession.Company),
		Start:   startTime,
		End:     endTime,
		Note:    repoSession.Note,
	}
}
