package cli

import (
	"fmt"
	"log"
	"os"

	"github.com/dormunis/punch/pkg/database"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [time]",
	Short: "Starts a new work day",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(companyName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp, err := getParsedTimeFromArgs(args)
		if err != nil {
			return err
		}

		day, err := timeTracker.StartDay(*company, timestamp, message)
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
		return getCompanyIfExists(companyName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		timestamp, err := getParsedTimeFromArgs(args)
		if err != nil {
			return err
		}

		day, err := timeTracker.EndDay(*company, timestamp, message)
		if err != nil {
			return err
		}
		printEOD(day)
		return nil
	},
}

func printBOD(day *database.Day) {
	fmt.Printf("Clocked in at %s\n", day.Start.Format("15:04:05"))
}

func printEOD(day *database.Day) {
	earnings, err := day.Earnings()
	duration := day.End.Sub(*day.Start)
	if err != nil {
		log.Fatalf("%v", err)
		os.Exit(1)
	}
	fmt.Printf("Clocked out at %s after %s (%.2f %s)\n",
		day.Start.Format("15:04:05"),
		duration,
		earnings,
		day.Company.Currency)
}

func init() {
	startCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	startCmd.Flags().StringVarP(&message, "message", "m", "", "Comment or message")
	endCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	endCmd.Flags().StringVarP(&message, "message", "m", "", "Comment or message")
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(endCmd)
}
