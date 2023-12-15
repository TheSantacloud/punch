package cli

import (
	"fmt"
	"log"

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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := getCompanyIfExists(currentCompanyName)
		if err != nil {
			log.Fatalf("%v", err)
		}
	},
}

var editCompanyCmd = &cobra.Command{
	Use:     "company [name]",
	Aliases: []string{"companies"},
	Short:   "edit a specific company",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		company, err := CompanyRepository.GetByName(args[0])
		if err != nil {
			log.Fatalf("%v", err)
		}
		buf, err := company.Serialize()
		if err != nil {
			log.Fatalf("%v", err)
		}
		err = editor.InteractiveEdit(buf, "yaml")
		if err != nil {
			log.Fatalf("%v", err)
		}
		var updateCompany models.Company
		err = models.DeserializeCompanyFromYAML(buf, &updateCompany)
		if err != nil {
			log.Fatalf("%v", err)
		}
		CompanyRepository.Update(&updateCompany)
		fmt.Printf("Updated company %s\n", updateCompany.Name)
	},
}

var editSessionCmd = &cobra.Command{
	Use:     "session [date]",
	Aliases: []string{"sessions"},
	Short:   "edit a specific session (defaults to latest today)",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var slice *[]models.Session
		var err error
		if allCompanies {
			slice, err = SessionRepository.GetAllSessionsAllCompanies()
			if err != nil {
				log.Fatalf("%v", err)
			}
		} else {
			timestamp, err := getParsedTimeFromArgs(args)
			if err != nil {
				log.Fatalf("%v", err)
			}
			session, err := SessionRepository.GetLatestSessionOnSpecificDate(timestamp, *currentCompany)
			if err != nil {
				log.Fatalf("%v", err)
			}
			slice = &[]models.Session{*session}
		}
		err = editSlice(slice)
		if err != nil {
			log.Fatalf("%v", err)
		}
		Puncher.Sync(slice)
		fmt.Printf("Updated %d session(s)\n", len(*slice))
	},
}

func editSlice(sessions *[]models.Session) error {
	buf, err := models.SerializeSessionsToYAML(*sessions)
	if err != nil {
		return err
	}

	err = editor.InteractiveEdit(buf, "yaml")
	if err != nil {
		return err
	}

	err = models.DeserializeSessionsFromYAML(buf, sessions)
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
