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
)

var rootCmd = &cobra.Command{
	Use:   "punch",
	Short: "Punchcard program for freelancers",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(companyName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		err := timeTracker.ToggleCheckInOut(company)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}

		day, err := timeTracker.GetDay(time.Now(), company)
		if day.End != nil {
			printEOD(day)
		} else {
			printBOD(day)
		}
	},
}

var startCmd = &cobra.Command{
	Use:   "start [time]",
	Short: "Starts a new work day",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(companyName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		timestamp, err := getParsedTimeFromArgs(args)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}

		day, err := timeTracker.StartDay(*company, timestamp)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}
		printBOD(day)

	},
}

var endCmd = &cobra.Command{
	Use:   "end [time]",
	Short: "End a work day",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(companyName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		timestamp, err := getParsedTimeFromArgs(args)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}

		day, err := timeTracker.EndDay(*company, timestamp)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}
		printEOD(day)
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
	rootCmd.Flags().StringP("message", "m", "", "Comment or message")

	startCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	endCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	endCmd.Flags().StringP("message", "m", "", "Comment or message")

	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(endCmd)
}

func Execute(cfg *config.Config) {
	var err error
	timeTracker, err = timetracker.NewTimeTracker(cfg)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		os.Exit(1)
	}

	err = rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
