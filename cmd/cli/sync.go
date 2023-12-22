package cli

import (
	"errors"
	"fmt"

	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/sync"
	"github.com/spf13/cobra"
)

var (
	pushOnly bool
	pullOnly bool
)

var syncCmd = &cobra.Command{
	Use:   "sync [remote]",
	Short: "sync sessions with remote",
	Args:  cobra.ExactArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if pullOnly && pushOnly {
			return errors.New("Cannot use both --pull-only and --push-only")
		}
		return nil
	},
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
			pullConflicts *[]models.Session
			pushConflicts *[]models.Session
		)

		if !pushOnly {
			pullConflicts, err = source.Pull()
			if err != nil {
				return err
			}

			if pullConflicts != nil && len(*pullConflicts) > 0 {
				fmt.Println("Pull Conflicts:")
				for _, conflict := range *pullConflicts {
					fmt.Println(conflict)
				}
			}
		}

		if !pullOnly {
			pushConflicts, err = source.Push()

			if err != nil {
				return err
			}

			if pushConflicts != nil && len(*pushConflicts) > 0 {
				fmt.Println("Push Conflicts:")
				for _, conflict := range *pushConflicts {
					fmt.Println(conflict)
				}
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	syncCmd.Flags().BoolVar(&pullOnly, "pull-only", false, "Only pull sessions from remote")
	syncCmd.Flags().BoolVar(&pushOnly, "push-only", false, "Only push sessions to remote")
}
