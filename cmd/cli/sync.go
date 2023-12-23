package cli

import (
	"errors"
	"fmt"

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

		var (
			pullConflicts map[models.Session]error
			pushConflicts map[models.Session]error
		)

		pullConflicts, err = source.Pull()
		if err != nil {
			return err
		}

		// TODO: create a conflict manager and use EditInteractive to resolve it
		if len(pullConflicts) > 0 {
			fmt.Println("Pull conflicts:")
			for session, err := range pullConflicts {
				fmt.Printf("Session ID: %v\nError: %v\n", session, err)
			}
		}

		if pullOnly || len(pullConflicts) > 0 {
			return nil
		}

		pushConflicts, err = source.Push()
		if err != nil {
			return err
		}

		if pushConflicts != nil && len(pushConflicts) > 0 {
			fmt.Println("Push conflicts:")
			for record, err := range pushConflicts {
				fmt.Printf("Record: %v\nError: %v\n", record, err)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&pullOnly, "pull-only", false, "Only pull sessions from remote")
}
