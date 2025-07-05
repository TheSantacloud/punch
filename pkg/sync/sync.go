package sync

import (
	"errors"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/dormunis/punch/pkg/sync/adapters/sheets"
)

type PushSummary struct {
	Added   int
	Updated int
	Errors  []error
}

type SyncSource interface {
	Type() string
	Pull() ([]models.Session, error)
	Push(*[]models.Session, *[]models.Session) (PushSummary, error)
}

var (
	ErrRemoteSourceNotSupported = errors.New("source not implemented")
	ErrInvalidRemoteConfig      = errors.New("invalid remote config")
	ErrSessionRepositoryNotSet  = errors.New("session repository not set")
)

func NewSource(remoteConfig config.Remote, sessionRepository repositories.SessionRepository) (SyncSource, error) {
	if sessionRepository == nil {
		return nil, ErrSessionRepositoryNotSet
	}

	switch remoteConfig.Type() {
	case "spreadsheet":
		remoteSheetConfig, ok := remoteConfig.(*config.SpreadsheetRemote)
		if !ok {
			return nil, ErrInvalidRemoteConfig
		}
		client, err := sheets.NewSheet(*remoteSheetConfig)
		if err != nil {
			return nil, err
		}
		return &SheetsSyncSource{
			Sheet:             client,
			SessionRepository: sessionRepository,
		}, nil
	default:
		return nil, ErrRemoteSourceNotSupported

	}
}
