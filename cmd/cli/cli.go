package cli

import (
	"os"
	"os/exec"

	"time"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/database"
	"github.com/dormunis/punch/pkg/timetracker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "open configuration file with your default editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		editor := viper.GetString("settings.editor")
		configFilePath := viper.ConfigFileUsed()
		command := exec.Command(editor, configFilePath)
		command.Stdin = os.Stdin
		command.Stdout = os.Stdout
		err := command.Run()
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	rootCmd.Flags().StringVarP(&message, "message", "m", "", "Comment or message")
	rootCmd.AddCommand(configCmd)
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
