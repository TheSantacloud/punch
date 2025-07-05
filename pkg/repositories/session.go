package repositories

import (
	"errors"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound = errors.New("session record not found")
	ErrConflictingIds  = errors.New("session already exists with a different ID")
	ErrInfoConflict    = errors.New("session exists with different info")
)

type RepoSession struct {
	ID         uint32 `gorm:"primaryKey;autoIncrement"`
	ClientName string `gorm:"foreignKey:Name"`
	Start      time.Time
	End        time.Time
	Note       string
	Client     RepoClient `gorm:"foreignKey:ClientName;references:Name"`
}

type GORMSessionRepository struct {
	db *gorm.DB
}

func NewGORMSessionRepository(db *gorm.DB) *GORMSessionRepository {
	return &GORMSessionRepository{db}
}

func (repo *GORMSessionRepository) Insert(session *models.Session, dryRun bool) error {
	repoSession := ToRepoSession(*session)

	sessionByID, err := repo.GetSessionByID(session.ID)
	if err == nil {
		if sessionByID.Conflicts(*session) {
			return ErrInfoConflict
		}
	}

	var existingByDetails RepoSession
	detailResult := repo.db.Where("start = ? AND client_name = ?",
		repoSession.Start,
		repoSession.ClientName).First(&existingByDetails)

	if detailResult.Error == nil {
		return ErrConflictingIds
	}

	if detailResult.Error == nil {
		sessionByDetails := ToDomainSession(existingByDetails)
		if sessionByDetails.Conflicts(*session) {
			return ErrInfoConflict
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

	var existingByDetails RepoSession
	detailResult := repo.db.Where("start = ? AND client_name = ?",
		repoSession.Start,
		repoSession.ClientName).First(&existingByDetails)

	if detailResult.Error != nil {
		session.ID = existingByDetails.ID
	}

	return repo.db.Save(&repoSession).Error
}

func (repo *GORMSessionRepository) GetSessionByID(id uint32) (*models.Session, error) {
	var repoSession RepoSession
	err := repo.db.Preload("Client").Where("id = ?", id).First(&repoSession).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	domainSession := ToDomainSession(repoSession)
	return &domainSession, nil
}

func (repo *GORMSessionRepository) GetLatestSession() (*models.Session, error) {
	var session RepoSession
	err := repo.db.Preload("Client").
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

func (repo *GORMSessionRepository) GetLatestSessionOnSpecificDateAllClients(date time.Time) (*[]models.Session, error) {
	var repoSessions []RepoSession
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := repo.db.Preload("Client").
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

func (repo *GORMSessionRepository) GetLatestSessionOnSpecificDate(date time.Time, client models.Client) (*models.Session, error) {
	var session RepoSession
	startOfDay := date.Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := repo.db.Preload("Client").
		Where("start >= ? AND start < ? AND client_name = ?",
			startOfDay,
			endOfDay,
			client.Name).
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

func (repo *GORMSessionRepository) GetAllSessions(client models.Client) (*[]models.Session, error) {
	var repoSessions []RepoSession
	err := repo.db.Preload("Client").
		Where("client_name = ?", client.Name).
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
	err := repo.db.Preload("Client").
		Where("start >= ?", start).
		Where(
			repo.db.Where("end < ?", end).
				Or("end IS NULL").
				Or("end IS ''")).
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

func (repo *GORMSessionRepository) GetLastSessions(count uint32, client *models.Client) (*[]models.Session, error) {
	var repoSessions []RepoSession
	var err error
	if client == nil {
		err = repo.db.Preload("Client").
			Order("start DESC").
			Limit(int(count)).
			Find(&repoSessions).Error
	} else {
		err = repo.db.Preload("Client").
			Where("client_name = ?", client.Name).
			Order("start DESC").
			Limit(int(count)).
			Find(&repoSessions).Error
	}

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

func (repo *GORMSessionRepository) GetAllSessionsAllClients() (*[]models.Session, error) {
	var repoSessions []RepoSession
	err := repo.db.Preload("Client").
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
	var clientName string
	if session.Client.Name != "" {
		clientName = session.Client.Name
	}

	startTime := session.Start.Truncate(time.Second)
	endTime := session.End
	if session.Finished() {
		endTime = endTime.Truncate(time.Second)
	}

	return RepoSession{
		ID:         session.ID,
		ClientName: clientName,
		Start:      startTime,
		End:        endTime,
		Note:       session.Note,
		Client:     *ToRepoClient(session.Client),
	}
}

func ToDomainSession(repoSession RepoSession) models.Session {
	startTime := repoSession.Start.In(time.Local)
	var endTime time.Time
	if repoSession.End != models.NULL_TIME {
		endTime = repoSession.End.In(time.Local)
	}
	return models.Session{
		ID:     repoSession.ID,
		Client: ToDomainClient(repoSession.Client),
		Start:  startTime,
		End:    endTime,
		Note:   repoSession.Note,
	}
}
