package cli

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [type]",
	Short: "Get a resource",
}

var getCompanyCmd = &cobra.Command{
	Use:     "company [name]",
	Short:   "Get a company",
	Aliases: []string{"companies"},
	PreRun: func(cmd *cobra.Command, args []string) {
		getCompanyIfExists(companyName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		if company != nil {
			company, err := timeTracker.GetCompany(company.Name)
			if err != nil {
				log.Fatalf("Unable to get company: %v", err)
			}
			fmt.Println(company.String())
		} else {
			companies, err := timeTracker.GetAllCompanies()
			if err != nil {
				log.Fatalf("Unable to get companies: %v", err)
			}
			for _, company := range *companies {
				fmt.Println(company.String())
			}
		}
	},
}

var getDayCmd = &cobra.Command{
	Use:   "day",
	Short: "Get a work day",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(companyName)
	},
	Run: func(cmd *cobra.Command, args []string) {
		today, err := timeTracker.GetDay(time.Now(), company)
		if err != nil {
			log.Fatalf("%v", err)
			os.Exit(1)
		}
		if verbose {
			fmt.Println(today.FullSummary())
			return
		}
		fmt.Println(today.Summary())
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getCompanyCmd)
	getCmd.AddCommand(getDayCmd)
	getDayCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	getCompanyCmd.Flags().StringVarP(&companyName, "company", "c", "", "Specify the company name")
	getDayCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
