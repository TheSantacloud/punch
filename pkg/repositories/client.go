package repositories

import (
	"errors"

	"github.com/dormunis/punch/pkg/models"
	"gorm.io/gorm"
)

var ErrClientNotFound = errors.New("record not found")

type RepoClient struct {
	Name     string `gorm:"primaryKey;collate:NOCASE"`
	PPH      uint16
	Currency string
}

type GORMClientRepository struct {
	db *gorm.DB
}

func NewGORMClientRepository(db *gorm.DB) *GORMClientRepository {
	return &GORMClientRepository{db}
}

func (repo *GORMClientRepository) GetAll() ([]models.Client, error) {
	var repoClients []RepoClient
	err := repo.db.Find(&repoClients).Error
	if err != nil {
		return nil, err
	}
	var clients []models.Client
	for _, repoClient := range repoClients {
		clients = append(clients, ToDomainClient(repoClient))
	}
	return clients, nil
}

func (repo *GORMClientRepository) Insert(client *models.Client) error {
	repoClient := ToRepoClient(*client)
	return repo.db.FirstOrCreate(repoClient, RepoClient{Name: repoClient.Name}).Error
}

func (repo *GORMClientRepository) Delete(client *models.Client) error {
	repoClient := ToRepoClient(*client)
	return repo.db.Delete(repoClient).Error
}

func (repo *GORMClientRepository) GetByName(name string) (*models.Client, error) {
	var repoClient RepoClient
	err := repo.db.Where("name = ?", name).First(&repoClient).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrClientNotFound
		}
		return nil, err
	}
	client := ToDomainClient(repoClient)
	return &client, nil
}

func (repo *GORMClientRepository) SafeGetByName(name string) (*models.Client, error) {
	var client RepoClient
	err := repo.db.Where("name = ?", name).First(&client).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	domainClient := ToDomainClient(client)
	return &domainClient, nil
}

func (repo *GORMClientRepository) Rename(client *models.Client, newName string) error {
	client.Name = newName
	repoClient := ToRepoClient(*client)
	return repo.db.Save(repoClient).Error
}

func (repo *GORMClientRepository) Update(client *models.Client) error {
	repoClient := ToRepoClient(*client)
	return repo.db.Save(repoClient).Error
}

func ToRepoClient(client models.Client) *RepoClient {
	return &RepoClient{
		Name:     client.Name,
		PPH:      client.PPH,
		Currency: client.Currency,
	}
}

func ToDomainClient(client RepoClient) models.Client {
	return models.Client{
		Name:     client.Name,
		PPH:      client.PPH,
		Currency: client.Currency,
	}
}
