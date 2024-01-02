package cli

import (
	"errors"

	"github.com/dormunis/punch/pkg/editor"
	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/sync"
	"github.com/spf13/cobra"
)

var (
	pullOnly bool
)

var syncCmd = &cobra.Command{
	Use:   "sync [remote]",
	Short: "sync sessions with remote",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var remoteString string
		Source = new(sync.SyncSource)
		if len(args) > 0 {
			remoteString = args[0]
		} else if len(Config.Settings.DefaultRemote) > 0 {
			remoteString = Config.Settings.DefaultRemote
		} else {
			return errors.New("Must specify remote")
		}
		remote, ok := Config.Remotes[remoteString]
		if !ok {
			return errors.New("Remote not found")
		}

		var err error
		*Source, err = sync.NewSource(remote, SessionRepository)
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return Sync()
	},
}

func Sync() error {
	err := pull(*Source)
	if err != nil {
		return err
	}

	if pullOnly {
		return nil
	}

	newSessions, err := SessionRepository.GetAllSessionsAllClients()
	if err != nil {
		return err
	}

	err = (*Source).Push(newSessions)
	if err != nil {
		return err
	}

	return nil
}

func pull(source sync.SyncSource) error {
	pulledSessions, err := source.Pull()
	if err != nil {
		return err
	}

	sessions, err := SessionRepository.GetAllSessionsAllClients()
	if err != nil {
		return err
	}

	conflictsBuffer, err := sync.GetConflicts(*sessions, pulledSessions)
	if err != nil {
		return err
	}

	if conflictsBuffer != nil && conflictsBuffer.Len() > 0 {
		err = editor.InteractiveEdit(conflictsBuffer, "yaml")
		if err != nil {
			return err
		}
	}

	deserializedSessions, err := models.DeserializeSessionsFromYAML(conflictsBuffer)
	if err != nil {
		return err
	}

	for _, session := range *deserializedSessions {
		err = SessionRepository.Upsert(&session, false)
		if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&pullOnly, "pull-only", false, "Only pull sessions from remote")
}
