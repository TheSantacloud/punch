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
	Push(*[]models.Session) (PushSummary, error)
}

var (
	RemoteSourceNotSupportedError = errors.New("Source not implemented")
	InvalidRemoteConfigError      = errors.New("Invalid remote config")
	SessionRepositoryNotSetError  = errors.New("Session repository not set")
)

func NewSource(remoteConfig config.Remote, sessionRepository repositories.SessionRepository) (SyncSource, error) {
	if sessionRepository == nil {
		return nil, SessionRepositoryNotSetError
	}

	switch remoteConfig.Type() {
	case "spreadsheet":
		remoteSheetConfig, ok := remoteConfig.(*config.SpreadsheetRemote)
		if !ok {
			return nil, InvalidRemoteConfigError
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
		return nil, RemoteSourceNotSupportedError

	}
}
