package cli

import (
	"fmt"
	"log"
	"strconv"

	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/spf13/cobra"
)

var (
	currency string
)

var addCmd = &cobra.Command{
	Use:   "add [type]",
	Short: "add a new resource",
}

var addCompanyCmd = &cobra.Command{
	Use:   "company [name] [price]",
	Short: "add a company",
	Long: `Add a company to the database. The price is the amount of money,  
    in your set currency, that the company pays you per hour.`,
	Args: cobra.ExactArgs(2),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if currency == "" && Config.Settings.Currency == "" {
			return fmt.Errorf("currency must be set")
		} else if currency == "" && Config.Settings.Currency != "" {
			currency = Config.Settings.Currency
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		price, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil || price <= 0 {
			log.Fatalf("invalid price %s", args[1])
		}
		newCompany := models.Company{
			Name:     name,
			PPH:      uint16(price),
			Currency: currency,
		}
		err = CompanyRepository.Insert(&newCompany)
		if err != nil && err != repositories.ErrCompanyNotFound {
			log.Fatalf("unable to insert company: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addCompanyCmd)
	addCompanyCmd.Flags().StringVar(&currency, "currency", "",
		"currency in which the company pays (defaults to USD)")
}
