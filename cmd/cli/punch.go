package cli

import (
	"fmt"

	"github.com/dormunis/punch/pkg/models"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [time]",
	Short: "Starts a new work session",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(currentCompanyName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp, err := getParsedTimeFromArgs(args)
		if err != nil {
			return err
		}

		session, err := Puncher.StartSession(*currentCompany, timestamp, punchMessage)
		if err != nil {
			return err
		}
		printBOD(session)
		return nil
	},
}

var endCmd = &cobra.Command{
	Use:   "end [time]",
	Short: "End a work session",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(currentCompanyName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp, err := getParsedTimeFromArgs(args)
		if err != nil {
			return err
		}

		session, _ := Puncher.EndSession(*currentCompany, timestamp, punchMessage)
		if err != nil {
			return err
		}
		printEOD(session)
		return nil
	},
}

func printBOD(session *models.Session) {
	fmt.Printf("Clocked in at %s\n", session.Start.Format("15:04:05"))
}

func printEOD(session *models.Session) error {
	earnings, err := session.Earnings()
	duration := session.End.Sub(*session.Start)
	if err != nil {
		return err
	}
	fmt.Printf("Clocked out at %s after %s (%.2f %s)\n",
		session.End.Format("15:04:05"),
		duration,
		earnings,
		session.Company.Currency)
	return nil
}

func init() {
	startCmd.Flags().StringVarP(&currentCompanyName, "company", "c", "", "Specify the company name")
	startCmd.Flags().StringVarP(&punchMessage, "message", "m", "", "Comment or message")
	endCmd.Flags().StringVarP(&currentCompanyName, "company", "c", "", "Specify the company name")
	endCmd.Flags().StringVarP(&punchMessage, "message", "m", "", "Comment or message")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(endCmd)
}
