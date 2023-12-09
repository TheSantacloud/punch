package cli

import (
	"fmt"

	"github.com/dormunis/punch/pkg/models"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [time]",
	Short: "Starts a new work day",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(currentCompanyName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp, err := getParsedTimeFromArgs(args)
		if err != nil {
			return err
		}

		day, err := Puncher.StartDay(*currentCompany, timestamp, punchMessage)
		if err != nil {
			return err
		}
		printBOD(day)
		return nil
	},
}

var endCmd = &cobra.Command{
	Use:   "end [time]",
	Short: "End a work day",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(currentCompanyName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp, err := getParsedTimeFromArgs(args)
		if err != nil {
			return err
		}

		day, _ := Puncher.EndDay(*currentCompany, timestamp, punchMessage)
		if err != nil {
			return err
		}
		printEOD(day)
		return nil
	},
}

func printBOD(day *models.Day) {
	fmt.Printf("Clocked in at %s\n", day.Start.Format("15:04:05"))
}

func printEOD(day *models.Day) error {
	earnings, err := day.Earnings()
	duration := day.End.Sub(*day.Start)
	if err != nil {
		return err
	}
	fmt.Printf("Clocked out at %s after %s (%.2f %s)\n",
		day.End.Format("15:04:05"),
		duration,
		earnings,
		day.Company.Currency)
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
