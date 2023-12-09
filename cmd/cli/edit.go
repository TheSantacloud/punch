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
	Short: "interactively edit work days",
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

var editDayCmd = &cobra.Command{
	Use:     "day [date]",
	Aliases: []string{"days"},
	Short:   "edit a specific day (default today)",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var slice *[]models.Day
		var err error
		if allCompanies {
			slice, err = DayRepository.GetAllDaysForAllCompanies()
			if err != nil {
				log.Fatalf("%v", err)
			}
		} else {
			timestamp, err := getParsedTimeFromArgs(args)
			if err != nil {
				log.Fatalf("%v", err)
			}
			day, err := DayRepository.GetDayFromDateForCompany(timestamp, *currentCompany)
			if err != nil {
				log.Fatalf("%v", err)
			}
			slice = &[]models.Day{*day}
		}
		err = editSlice(slice)
		if err != nil {
			log.Fatalf("%v", err)
		}
		Puncher.Sync(slice)
		fmt.Printf("Updated %d day(s)\n", len(*slice))
	},
}

func editSlice(days *[]models.Day) error {
	buf, err := models.SerializeDaysToYAML(*days)
	if err != nil {
		return err
	}

	err = editor.InteractiveEdit(buf, "yaml")
	if err != nil {
		return err
	}

	err = models.DeserializeDaysFromYAML(buf, days)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.AddCommand(editDayCmd)
	editCmd.AddCommand(editCompanyCmd)
	editDayCmd.Flags().StringVarP(&currentCompanyName, "company", "c", "", "Specify the company name")
	editDayCmd.Flags().BoolVarP(&allCompanies, "all", "a", false, "Edit all companies")
}
