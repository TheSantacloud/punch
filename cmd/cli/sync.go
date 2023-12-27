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
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		remote, ok := Config.Remotes[args[0]]
		if !ok {
			return errors.New("Remote not found")
		}

		source, err := sync.NewSource(remote, SessionRepository)
		if err != nil {
			return err
		}

		resolvedSessions, err := pull(source)
		if err != nil {
			return err
		}

		if pullOnly {
			return nil
		}

		err = source.Push(resolvedSessions)
		if err != nil {
			return err
		}

		return nil
	},
}

func pull(source sync.SyncSource) (*[]models.Session, error) {
	pulledSessions, err := source.Pull()
	if err != nil {
		return nil, err
	}

	sessions, err := SessionRepository.GetAllSessionsAllCompanies()
	if err != nil {
		return nil, err
	}

	conflictsBuffer, err := sync.GetConflicts(*sessions, pulledSessions)
	if err != nil {
		return nil, err
	}

	if conflictsBuffer != nil && conflictsBuffer.Len() > 0 {
		err = editor.InteractiveEdit(conflictsBuffer, "yaml")
		if err != nil {
			return nil, err
		}
	}

	err = models.DeserializeSessionsFromYAML(conflictsBuffer, sessions, false)
	if err != nil {
		return nil, err
	}

	for _, session := range *sessions {
		err = SessionRepository.Upsert(&session, false)
		if err != nil {
			return nil, err
		}
	}

	return sessions, nil
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&pullOnly, "pull-only", false, "Only pull sessions from remote")
}
