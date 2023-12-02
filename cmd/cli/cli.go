package cli

import (
	"os"

	"fmt"

	"log"
	"time"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/database"
	"github.com/dormunis/punch/pkg/timetracker"
	"github.com/spf13/cobra"
)

var (
	timeTracker *timetracker.TimeTracker
	companyName string
	company     *database.Company
	verbose     bool
	message     string
)

var rootCmd = &cobra.Command{
	Use:   "punch",
	Short: "Punchcard program for freelancers",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(companyName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		err := timeTracker.ToggleCheckInOut(company, message)
		if err != nil {
			return err
		}

		day, err := timeTracker.GetDay(time.Now(), company)
		if day.End != nil {
			printEOD(day)
		} else {
			printBOD(day)
		}
		return nil
	},
}

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
	rootCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	startCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	endCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")

	rootCmd.Flags().StringVarP(&message, "message", "m", "", "Comment or message")
	startCmd.Flags().StringVarP(&message, "message", "m", "", "Comment or message")
	endCmd.Flags().StringVarP(&message, "message", "m", "", "Comment or message")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(endCmd)
}

func Execute(cfg *config.Config) error {
	var err error
	timeTracker, err = timetracker.NewTimeTracker(cfg)
	if err != nil {
		return err
	}

	err = rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}
