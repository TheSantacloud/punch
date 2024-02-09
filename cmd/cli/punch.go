package cli

import (
	"fmt"
	"slices"

	"github.com/dormunis/punch/pkg/models"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [time]",
	Short: "Starts a new work session",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return GetClientIfExists(currentClientName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp, _, err := ExtractParsedTimeFromArgs(args, currentClientName)
		if err != nil {
			return err
		}

		session, err := Puncher.StartSession(*currentClient, timestamp, punchMessage)
		if err != nil {
			return err
		}
		printBOD(cmd, session)

		if slices.Contains(Config.Settings.AutoSync, "start") {
			err = Sync(rootCmd)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

var endCmd = &cobra.Command{
	Use:   "end [time]",
	Short: "End a work session",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return GetClientIfExists(currentClientName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp, _, err := ExtractParsedTimeFromArgs(args, currentClientName)
		if err != nil {
			return err
		}

		currentSession, err := SessionRepository.GetLatestSession()
		if err != nil {
			return err
		}

		session, _ := Puncher.EndSession(*currentSession, timestamp, punchMessage)
		if err != nil {
			return err
		}
		err = printEOD(cmd, session)
		if err != nil {
			return err
		}

		if slices.Contains(Config.Settings.AutoSync, "end") {
			err = Sync(rootCmd)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func printBOD(cmd *cobra.Command, session *models.Session) {
	fmt.Printf("Clocked in at %s\n", session.Start.Format("15:04:05"))
}

func printEOD(cmd *cobra.Command, session *models.Session) error {
	earnings, err := session.Earnings()
	duration := session.End.Sub(*session.Start)
	if err != nil {
		return err
	}
	cmd.Printf("Clocked out at %s after %s (%.2f %s)\n",
		session.End.Format("15:04:05"),
		duration,
		earnings,
		session.Client.Currency)
	return nil
}

func init() {
	startCmd.Flags().StringVarP(&currentClientName, "client", "c", "", "Specify the client name")
	startCmd.Flags().StringVarP(&punchMessage, "message", "m", "", "Comment or message")
	endCmd.Flags().StringVarP(&currentClientName, "client", "c", "", "Specify the client name")
	endCmd.Flags().StringVarP(&punchMessage, "message", "m", "", "Comment or message")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(endCmd)
}
