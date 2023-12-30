package cli

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dormunis/punch/pkg/editor"
	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/sync"
	"github.com/spf13/cobra"
)

var (
	allClients    bool
	dateString    string
	approveDelete bool
)

var editCmd = &cobra.Command{
	Use:   "edit [time]",
	Short: "interactively edit work sessions",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := getClientIfExists(currentClientName)
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
		fmt.Printf("Updated client %s\n", updateClient.Name)
		return nil
	},
}

var editSessionCmd = &cobra.Command{
	Use:     "session [id]",
	Aliases: []string{"sessions"},
	Short:   "edit a specific session (defaults to latest today)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if (allClients && len(args) > 0) ||
			(allClients && dateString != "") ||
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
		if len(deletedSessions) > 0 && verifyDeletion(deletedSessions) {
			for _, session := range deletedSessions {
				err = SessionRepository.Delete(&session, false)
				if err != nil {
					fmt.Printf("Unable to delete session %s: %v\n", session.String(), err)
				}
			}
			fmt.Printf("Deleted %d session(s)\n", len(deletedSessions))
		}

		if slices.Contains(Config.Settings.AutoSync, "edit") {
			err = Sync()
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func getSessionSlice(args []string) (*[]models.Session, error) {
	var slice *[]models.Session
	var err error

	if allClients {
		slice, err = SessionRepository.GetAllSessionsAllClients()
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
		slice, err = SessionRepository.GetAllSessionsBetweenDates(startOfDay, endOfDay)
		if err != nil {
			return nil, err
		}
	} else {
		today := time.Now()
		session, err := SessionRepository.GetLatestSessionOnSpecificDate(today, *currentClient)
		if err != nil {
			return nil, err
		}
		slice = &[]models.Session{*session}
	}

	if slice != nil && len(*slice) == 0 {
		return nil, errors.New("no sessions found")
	}

	sort.SliceStable(*slice, func(i, j int) bool {
		return (*slice)[i].Start.Before(*(*slice)[j].Start)
	})

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
	editCmd.AddCommand(editClientCmd)
	editSessionCmd.Flags().StringVarP(&currentClientName, "client", "c", "", "Specify the client name")
	editSessionCmd.Flags().StringVarP(&dateString, "date", "d", "", "Specify a specific date to edit sessions for")
	editSessionCmd.Flags().BoolVarP(&allClients, "all", "a", false, "Edit all clients")
	editSessionCmd.Flags().BoolVarP(&approveDelete, "yes", "y", false, "Approve deletion of sessions automatically")
}
