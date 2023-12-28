package cli

import (
	"fmt"
	"slices"
	"time"

	"github.com/dormunis/punch/pkg/editor"
	"github.com/dormunis/punch/pkg/models"
	"github.com/spf13/cobra"
)

var (
	allCompanies bool
)

var editCmd = &cobra.Command{
	Use:   "edit [time]",
	Short: "interactively edit work sessions",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getCompanyIfExists(currentCompanyName)
		if err != nil {
			return err
		}
		return nil
	},
}

var editCompanyCmd = &cobra.Command{
	Use:     "company [name]",
	Aliases: []string{"companies"},
	Short:   "edit a specific company",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: edit name should edit rather than add
		// TODO: delete functionality
		company, err := CompanyRepository.GetByName(args[0])
		if err != nil {
			return err
		}
		buf, err := company.Serialize()
		if err != nil {
			return err
		}
		err = editor.InteractiveEdit(buf, "yaml")
		if err != nil {
			return err
		}
		var updateCompany models.Company
		err = models.DeserializeCompanyFromYAML(buf, &updateCompany)
		if err != nil {
			return err
		}
		CompanyRepository.Update(&updateCompany)
		fmt.Printf("Updated company %s\n", updateCompany.Name)
		return nil
	},
}

var editSessionCmd = &cobra.Command{
	Use:     "session [date]",
	Aliases: []string{"sessions"},
	Short:   "edit a specific session (defaults to latest today)",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var slice *[]models.Session
		var err error
		if allCompanies {
			slice, err = SessionRepository.GetAllSessionsAllCompanies()
			if err != nil {
				return err
			}
		} else {
			timestamp, err := getParsedTimeFromArgs(args)
			if err != nil {
				return err
			}
			startOfDay := timestamp.Truncate(24 * time.Hour)
			endOfDay := startOfDay.Add(24 * time.Hour)
			slice, err = SessionRepository.GetAllSessionsBetweenDates(*currentCompany, startOfDay, endOfDay)
		}
		if len(*slice) == 0 {
			fmt.Println("No sessions found")
			return nil
		}

		err = editSessionSlice(slice)
		if err != nil {
			return err
		}
		sessionsUpdatedCount := 0
		for _, session := range *slice {
			err = SessionRepository.Update(&session, false)
			if err != nil {
				fmt.Printf("Unable to update session %s: %v\n", session.Start, err)
			} else {
				sessionsUpdatedCount++
			}
		}
		fmt.Printf("Updated %d session(s)\n", sessionsUpdatedCount)

		fmt.Println(Config.Settings.AutoSync)
		if slices.Contains(Config.Settings.AutoSync, "edit") {
			Sync()
		}
		return nil
	},
}

func editSessionSlice(sessions *[]models.Session) error {
	buf, err := models.SerializeSessionsToYAML(*sessions)
	if err != nil {
		return err
	}

	err = editor.InteractiveEdit(buf, "yaml")
	if err != nil {
		return err
	}

	err = models.DeserializeSessionsFromYAML(buf, sessions, true)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editSessionCmd)
	editCmd.AddCommand(editCompanyCmd)
	editSessionCmd.Flags().StringVarP(&currentCompanyName, "company", "c", "", "Specify the company name")
	editSessionCmd.Flags().BoolVarP(&allCompanies, "all", "a", false, "Edit all companies")
}
