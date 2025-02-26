package cli

import (
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
	approveDelete bool
)

var editCmd = &cobra.Command{
	Use:   "edit [time]",
	Short: "interactively edit work sessions",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := GetClientIfExists(currentClientName)
		if err != nil {
			return err
		}
		return nil
	},
}

var editClientCmd = &cobra.Command{
	Use:     "client [name]",
	Aliases: []string{"clients"},
	Short:   "edit a specific client",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := ClientRepository.GetByName(args[0])
		if err != nil {
			return err
		}
		buf, err := client.Serialize()
		if err != nil {
			return err
		}
		err = editor.InteractiveEdit(buf, "yaml")
		if err != nil {
			return err
		}
		var updateClient models.Client
		err = models.DeserializeClientFromYAML(buf, &updateClient)
		if err != nil {
			return err
		}
		err = ClientRepository.Update(&updateClient)
		if err != nil {
			return err
		}
		rootCmd.Printf("Updated client %s\n", updateClient.Name)
		return nil
	},
}

var editSessionCmd = &cobra.Command{
	Use:     "session [id]",
	Aliases: []string{"sessions"},
	Short:   "edit a specific session (defaults to latest today)",
	Args:    cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		reportTimeframe, err = ExtractTimeframeFromFlags()
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var sessions []models.Session
		if len(args) == 0 {
			sessions = GetSessionsWithTimeframe(*reportTimeframe)
		} else {
			sessions = GetRelativeSessionsFromArgs(args, clientName)
		}

		SortSessions(&sessions, descendingOrder)

		editedSessions, err := editSessionSlice(&sessions)
		if err != nil {
			return err
		}

		sessionsUpdatedCount := 0
		for _, session := range editedSessions {
			for _, previousSession := range sessions {
				if session.ID == previousSession.ID && !previousSession.Equals(session) {
					sessionsUpdatedCount++
					err = SessionRepository.Update(&session, false)
					if err != nil {
						rootCmd.Printf("Unable to update session %s: %v\n", session.Start, err)
					}
				}
			}
		}

		rootCmd.Printf("Updated %d session(s)\n", sessionsUpdatedCount)

		deletedSessions := sync.DetectDeletedSessions(&sessions, &editedSessions)
		if len(deletedSessions) > 0 && verifyDeletion(deletedSessions) {
			for _, session := range deletedSessions {
				err = SessionRepository.Delete(&session, false)
				if err != nil {
					rootCmd.Printf("Unable to delete session %s: %v\n", session.String(), err)
				}
			}
			rootCmd.Printf("Deleted %d session(s)\n", len(deletedSessions))
		}

		if slices.Contains(Config.Settings.AutoSync, "edit") {
			err = Sync(rootCmd)
			if err != nil {
				return err
			}
		}
		return nil
	},
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

	rootCmd.Printf("Detected %d deleted session(s)\n", len(deleted))
	rootCmd.Print("Are you sure you want to delete these sessions (y/n)? ")
	var answer string
	decision, err := fmt.Scanln(&answer)
	if decision != 1 || err != nil {
		return false
	}
	return strings.ToLower(answer) == "y"

}

func init() {
	currentYear, currentMonth, _ := time.Now().Date()
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editSessionCmd)
	editCmd.AddCommand(editClientCmd)
	editSessionCmd.Flags().StringVarP(&clientName, "client", "c", "", "Specify the client name")
	editSessionCmd.Flags().BoolVar(&dayReport, "day", false, "Edit report for this current day")
	editSessionCmd.Flags().BoolVar(&weekReport, "week", false, "Edit report for this current week")
	editSessionCmd.Flags().StringVar(&monthReport, "month", "", "Edit report for a specific month (format: YYYY-MM), leave empty for current month")
	editSessionCmd.Flags().StringVar(&yearReport, "year", "", "Edit report for a specific year (format: YYYY), leave empty for current year")
	editSessionCmd.Flags().BoolVarP(&allReport, "all", "a", false, "Edit all clients")
	editSessionCmd.Flags().BoolVarP(&approveDelete, "yes", "y", false, "Approve deletion of sessions automatically")
	editSessionCmd.Flags().Lookup("month").NoOptDefVal = strconv.Itoa(int(currentMonth))
	editSessionCmd.Flags().Lookup("year").NoOptDefVal = strconv.Itoa(currentYear)
}
