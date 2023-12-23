package sync

import (
	"errors"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/dormunis/punch/pkg/sync/adapters/sheets"
)

type SyncSource interface {
	Type() string
	Pull() (map[models.Session]error, error)
	Push() (map[models.Session]error, error)
}

var (
	RemoteSourceNotSupportedError = errors.New("Source not implemented")
	InvalidRemoteConfigError      = errors.New("Invalid remote config")
	SessionRepositoryNotSetError  = errors.New("Session repository not set")
	PullConflictError             = errors.New("Pull conflict error")
	PushConflictError             = errors.New("Push conflict error")
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
		client, err := sheets.GetSheet(*remoteSheetConfig)
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
