package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/database"
	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/puncher"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// common instances
var (
	Config            *config.Config
	DayRepository     repositories.DayRepository
	CompanyRepository repositories.CompanyRepository
	Puncher           *puncher.Puncher
)

// cli flags
var (
	currentCompanyName string
	currentCompany     *models.Company
	punchMessage       string
	verbose            bool
)

var rootCmd = &cobra.Command{
	Use:           "punch",
	Short:         "Punchcard program for freelancers",
	SilenceUsage:  true,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(currentCompanyName)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		day, err := Puncher.ToggleCheckInOut(currentCompany, punchMessage)
		if err != nil {
			return err
		}

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
	rootCmd.Flags().StringVarP(&currentCompanyName, "company", "c", "", "Specify the company name")
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

	DayRepository = repositories.NewGORMDayRepository(db)
	CompanyRepository = repositories.NewGORMCompanyRepository(db)
	Puncher = puncher.NewPuncher(DayRepository)

	err = rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}
