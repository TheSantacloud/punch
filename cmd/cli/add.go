package cli

import (
	"log"
	"strconv"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use: "add [type]",
}

var addCompanyCmd = &cobra.Command{
	Use:   "company [name] [price]",
	Short: "add a company",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if getCompanyIfExists(args[0]) == nil {
			log.Fatalf("company %s already exists", args[0])
		}
		name := args[0]
		price, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil || price <= 0 {
			log.Fatalf("invalid price %s", args[1])
		}
		timeTracker.AddCompany(name, int32(price))
	},
}

var addDayCmd = &cobra.Command{
	Use:   "day [date]",
	Short: "add a work day",
	Args:  cobra.MaximumNArgs(1),
	PreRunE: func(cmd *cobra.Command, args []string) error {
		return getCompanyIfExists(companyName)
	},
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addCompanyCmd)
	addCmd.AddCommand(addDayCmd)
}
