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

var addClientCmd = &cobra.Command{
	Use:   "client [name] [price]",
	Short: "add a client",
	Long: `Add a client to the database. The price is the amount of money,  
    in your set currency, that the client pays you per hour.`,
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
		if err != nil || price < 0 {
			log.Fatalf("invalid price %s", args[1])
		}
		newClient := models.Client{
			Name:     name,
			PPH:      uint16(price),
			Currency: currency,
		}
		err = ClientRepository.Insert(&newClient)
		if err != nil && err != repositories.ErrClientNotFound {
			log.Fatalf("unable to insert client: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
	addCmd.AddCommand(addClientCmd)
	addClientCmd.Flags().StringVar(&currency, "currency", "",
		"currency in which the client pays (defaults to USD)")
}
