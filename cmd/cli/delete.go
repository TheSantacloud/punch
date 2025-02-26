package cli

import (
	"slices"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "delete command",
}

var deleteClientCmd = &cobra.Command{
	Use:     "client [name]",
	Aliases: []string{"clients"},
	Short:   "delete a specific client",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := ClientRepository.GetByName(args[0])
		if err != nil {
			return err
		}
		return ClientRepository.Delete(client)
	},
}

var deleteSessionCmd = &cobra.Command{
	Use:     "session [id]",
	Aliases: []string{"sessions"},
	Short:   "delete a specific session by id",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := GetSessionByID(args[0])
		if err != nil {
			return err
		}

		err = SessionRepository.Delete(session, false)
		if err != nil {
			return err
		}

		rootCmd.Printf("Deleted session (%d) %s\n", session.ID, session.String())

		if slices.Contains(Config.Settings.AutoSync, "delete") {
			err = Sync(rootCmd)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteSessionCmd)
	deleteCmd.AddCommand(deleteClientCmd)
}
