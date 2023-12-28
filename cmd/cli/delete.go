package cli

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [id]",
	Short: "delete command",
}

var deleteCompanyCmd = &cobra.Command{
	Use:     "company [name]",
	Aliases: []string{"companies"},
	Short:   "delete a specific company",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		company, err := CompanyRepository.GetByName(args[0])
		if err != nil {
			return err
		}
		return CompanyRepository.Delete(company)
	},
}

var deleteSessionCmd = &cobra.Command{
	Use:     "session [id]",
	Aliases: []string{"sessions"},
	Short:   "delete a specific session by id",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionId, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return err
		}

		session, err := SessionRepository.GetSessionByID(uint32(sessionId))
		if err != nil {
			return err
		}

		err = SessionRepository.Delete(session, false)
		if err != nil {
			return err
		}

		fmt.Printf("Deleted session (%d) %s\n", sessionId, session.String())

		if slices.Contains(Config.Settings.AutoSync, "delete") {
			Sync()
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.AddCommand(deleteSessionCmd)
	deleteCmd.AddCommand(deleteCompanyCmd)
}
