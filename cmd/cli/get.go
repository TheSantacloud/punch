package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/spf13/cobra"
)

var (
	output          string
	descendingOrder bool
)

var getCmd = &cobra.Command{
	Use:   "get [type]",
	Short: "Get a resource",
}

var getClientCmd = &cobra.Command{
	Use:     "client [name]",
	Short:   "Get a client",
	Aliases: []string{"clients"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 {
			client, err := ClientRepository.GetByName(args[0])
			if err != nil {
				return fmt.Errorf("Unable to get client: %v", err)
			}
			fmt.Println(client.String())
		} else {
			clients, err := ClientRepository.GetAll()
			if err != nil {
				return fmt.Errorf("Unable to get clients: %v", err)
			}
			for _, client := range clients {
				fmt.Println(client.String())
			}
		}
		return nil
	},
}

var getSessionCmd = &cobra.Command{
	Use:   "session [date]",
	Short: "Get a work session",
	Long: `Get a work session. If no date is specified, the latest of current day is used.
    If a date is specified, the format must be YYYY-MM-DD.`,
	Example: `punch get session 
punch get session 2020-01-01
punch get session 01-01`,
	Aliases: []string{"sessions"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := preRunCheckOutput()
		if err != nil {
			return err
		}
		reportTimeframe, err = GetTimeframe()
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		slice, err := GetRelevatSessions(*reportTimeframe)
		if err != nil {
			return err
		}

		SortSessions(slice, descendingOrder)

		content, err := generateView(slice)
		if err != nil {
			return err
		}
		fmt.Print(*content)
		return nil
	},
}

func generateView(slice *[]models.Session) (*string, error) {
	if len(*slice) == 0 {
		return nil, errors.New("No available data")
	}
	var (
		buffer        *bytes.Buffer
		err           error
		content       string        = ""
		totalDuration time.Duration = 0
		totalEarnings float64       = 0
		clientName    string        = ""
		currency      string        = ""
	)

	switch output {
	case "text":
		lastSessionDate := (*slice)[0].Start
		for _, session := range *slice {
			if verbose {
				content += session.Summary()
				content += "\n"
			}
			if session.End == nil {
				continue
			}
			totalDuration += session.End.Sub(*session.Start)
			earnings, _ := session.Earnings()
			totalEarnings += earnings
			// TODO: dont get just the latest client, figure out a better print here
			clientName = session.Client.Name
			currency = session.Client.Currency

			if session.Start.Day() > lastSessionDate.Day() {
				lastSessionDate = session.Start
			}

		}

		content += fmt.Sprintf("%s\t%s\t%s\t%.2f %s\n",
			lastSessionDate.Format("2006-01-02"),
			clientName,
			totalDuration,
			totalEarnings,
			currency,
		)
	case "csv":
		if verbose {
			buffer, err = models.SerializeSessionsToFullCSV(*slice)
		} else {
			buffer, err = models.SerializeSessionsToCSV(*slice)
		}
		if err != nil {
			return nil, err
		}
		content = buffer.String()
	}
	return &content, nil
}

func preRunCheckOutput() error {
	allowedOutputs := map[string]bool{"csv": true, "text": true}
	if _, ok := allowedOutputs[output]; !ok {
		return fmt.Errorf("invalid output format: %s, allowed formats are 'csv' and 'text'", output)
	}
	return nil

}

func init() {
	currentYear, currentMonth, _ := time.Now().Date()
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getSessionCmd)
	getCmd.AddCommand(getClientCmd)
	getSessionCmd.Flags().StringVarP(&currentClientName, "client", "c", "", "Specify the client name")
	getSessionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	getSessionCmd.Flags().BoolVar(&dayReport, "day", false, "Get report for this current day")
	getSessionCmd.Flags().BoolVar(&weekReport, "week", false, "Get report for this current week")
	getSessionCmd.Flags().StringVar(&monthReport, "month", "", "Get report for a specific month (format: YYYY-MM), leave empty for current month")
	getSessionCmd.Flags().StringVar(&yearReport, "year", "", "Get report for a specific year (format: YYYY), leave empty for current year")
	getSessionCmd.Flags().BoolVar(&allReport, "all", false, "Get all sessions")
	getSessionCmd.Flags().BoolVar(&descendingOrder, "desc", false, "Sort sessions in descending order (defaults to ascending order)")
	getSessionCmd.Flags().StringVarP(&output, "output", "o", "text", "Specify the output format")
	getSessionCmd.Flags().Lookup("month").NoOptDefVal = strconv.Itoa(int(currentMonth))
	getSessionCmd.Flags().Lookup("year").NoOptDefVal = strconv.Itoa(currentYear)
}
