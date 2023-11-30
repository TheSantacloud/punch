package cli

import (
	"log"

	"github.com/dormunis/punch/pkg/database"
	"github.com/dormunis/punch/pkg/editor"
	"github.com/spf13/cobra"
)

var (
	allCompanies bool
)

var editCmd = &cobra.Command{
	Use:   "edit [time]",
	Short: "interactively edit work days",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := getCompanyIfExists(companyName)
		if err != nil {
			log.Fatalf("%v", err)
		}
	},
}

var editDayCmd = &cobra.Command{
	Use:     "day [date]",
	Aliases: []string{"days"},
	Short:   "edit a specific day (default today)",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var slice *[]database.Day
		var err error
		if allCompanies {
			slice, err = timeTracker.GetAllDays(company)
			if err != nil {
				log.Fatalf("%v", err)
			}
		} else {
			timestamp, err := getParsedTimeFromArgs(args)
			if err != nil {
				log.Fatalf("%v", err)
			}
			day, err := timeTracker.GetDay(timestamp, company)
			if err != nil {
				log.Fatalf("%v", err)
			}
			slice = &[]database.Day{*day}
		}
		editSlice(slice)
	},
}

func editSlice(dailies *[]database.Day) error {
	buf, err := database.SerializeDaysToYAML(*dailies)
	if err != nil {
		return err
	}

	err = editor.InteractiveEdit(buf, "yaml")
	if err != nil {
		return err
	}

	err = database.DeserializeDaysFromYAML(buf, dailies)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editDayCmd)
	editDayCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	editDayCmd.Flags().BoolVarP(&allCompanies, "all", "a", false, "Edit all companies")
}
