package cli

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/dormunis/punch/pkg/editor"
	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/sync"
	"github.com/spf13/cobra"
)

var (
	allCompanies  bool
	dateString    string
	approveDelete bool
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
	Use:     "session [id]",
	Aliases: []string{"sessions"},
	Short:   "edit a specific session (defaults to latest today)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if (allCompanies && len(args) > 0) ||
			(allCompanies && dateString != "") ||
			(len(args) > 0 && dateString != "") {
			return fmt.Errorf("Cannot specify both --all and --date or session id")
		}

		sessions, err := getSessionSlice(args)
		if err != nil {
			return err
		}

		editedSessions, err := editSessionSlice(sessions)
		if err != nil {
			return err
		}

		sessionsUpdatedCount := 0
		for _, session := range editedSessions {
			err = SessionRepository.Update(&session, false)
			if err != nil {
				fmt.Printf("Unable to update session %s: %v\n", session.Start, err)
			} else {
				sessionsUpdatedCount++
			}
		}
		fmt.Printf("Updated %d session(s)\n", sessionsUpdatedCount)

		deletedSessions := sync.DetectDeletedSessions(sessions, &editedSessions)
		if verifyDeletion(deletedSessions) {
			for _, session := range deletedSessions {
				err = SessionRepository.Delete(&session, false)
				if err != nil {
					fmt.Printf("Unable to delete session %s: %v\n", session.String(), err)
				}
			}
		}
		fmt.Printf("Deleted %d session(s)\n", len(deletedSessions))

		if slices.Contains(Config.Settings.AutoSync, "edit") {
			Sync()
		}
		return nil
	},
}

func getSessionSlice(args []string) (*[]models.Session, error) {
	var slice *[]models.Session
	var err error

	if allCompanies {
		slice, err = SessionRepository.GetAllSessionsAllCompanies()
		if err != nil {
			return nil, err
		}
	} else if len(args) > 0 {
		sessionId, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return nil, err
		}
		session, err := SessionRepository.GetSessionByID(uint32(sessionId))
		if err != nil {
			return nil, err
		}
		slice = &[]models.Session{*session}
	} else if dateString != "" {
		timestamp, err := getParsedTimeFromArgs([]string{dateString})
		if err != nil {
			return nil, err
		}
		startOfDay := timestamp.Truncate(24 * time.Hour)
		endOfDay := startOfDay.Add(24 * time.Hour)
		slice, err = SessionRepository.GetAllSessionsBetweenDates(*currentCompany, startOfDay, endOfDay)
	} else {
		today := time.Now().Truncate(24 * time.Hour)
		session, err := SessionRepository.GetLatestSessionOnSpecificDate(today, *currentCompany)
		if err != nil {
			return nil, err
		}
		slice = &[]models.Session{*session}
	}

	if slice != nil && len(*slice) == 0 {
		return nil, errors.New("no sessions found")
	}
	return slice, nil
}

func editSessionSlice(sessions *[]models.Session) ([]models.Session, error) {
	buf, err := models.SerializeSessionsToYAML(*sessions)
	if err != nil {
		return nil, err
	}

	err = editor.InteractiveEdit(buf, "yaml")
	if err != nil {
		return nil, err
	}

	deserializedSessions, err := models.DeserializeSessionsFromYAML(buf)
	if err != nil {
		return nil, err
	}

	return *deserializedSessions, nil
}

func verifyDeletion(deleted []models.Session) bool {
	if approveDelete {
		return true
	}

	fmt.Printf("Detected %d deleted session(s)\n", len(deleted))
	fmt.Print("Are you sure you want to delete these sessions (y/n)? ")
	var answer string
	fmt.Scanln(&answer)

	return strings.ToLower(answer) == "y"

}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editSessionCmd)
	editCmd.AddCommand(editCompanyCmd)
	editSessionCmd.Flags().StringVarP(&currentCompanyName, "company", "c", "", "Specify the company name")
	editSessionCmd.Flags().StringVarP(&dateString, "date", "d", "", "Specify a specific date to edit sessions for")
	editSessionCmd.Flags().BoolVarP(&allCompanies, "all", "a", false, "Edit all companies")
	editSessionCmd.Flags().BoolVarP(&approveDelete, "yes", "y", false, "Approve deletion of sessions automatically")
}
