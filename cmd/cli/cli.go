package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/database"
	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/puncher"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/dormunis/punch/pkg/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// common instances
var (
	Config            *config.Config
	SessionRepository repositories.SessionRepository
	ClientRepository  repositories.ClientRepository
	Puncher           *puncher.Puncher
	Source            *sync.SyncSource
)

// cli flags
var (
	currentClientName string
	currentClient     *models.Client
	punchMessage      string
	verbose           bool
)

var rootCmd = &cobra.Command{
	Use:           "punch",
	Short:         "Punchcard program for freelancers",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getClientIfExists(currentClientName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		session, err := Puncher.ToggleCheckInOut(currentClient, punchMessage)
		if err != nil {
			return err
		}

		if session.End != nil {
			printEOD(session)
			if slices.Contains(Config.Settings.AutoSync, "end") {
				Sync()
			}
		} else {
			printBOD(session)
			if slices.Contains(Config.Settings.AutoSync, "start") {
				Sync()
			}
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
	rootCmd.Flags().StringVarP(&currentClientName, "client", "c", "", "Specify a client's name")
	rootCmd.Flags().StringVarP(&punchMessage, "message", "m", "", "Comment or message")
	rootCmd.AddCommand(configCmd)
}

func Execute(cfg *config.Config) error {
	var err error
	Config = cfg

	db, err := database.NewDatabase(cfg.Database.Engine, cfg.Database.Path)
	if err != nil {
		return fmt.Errorf("Unable to connect to models. %v", err)
	}

	SessionRepository = repositories.NewGORMSessionRepository(db)
	ClientRepository = repositories.NewGORMClientRepository(db)
	Puncher = puncher.NewPuncher(SessionRepository)

	if Config.Settings.DefaultRemote != "" {
		remote, ok := Config.Remotes[Config.Settings.DefaultRemote]
		if !ok {
			return errors.New("Remote not found")
		}
		Source = new(sync.SyncSource)
		*Source, err = sync.NewSource(remote, SessionRepository)
		if err != nil {
			return err
		}
	} else {
		Source = nil
	}

	err = rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}
