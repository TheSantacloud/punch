package cli

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/spf13/cobra"
)

var (
	output          string
	descendingOrder bool
	summary         bool
	hideHeaders     bool

	ErrNoAvailableData = errors.New("no available data")
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
				return fmt.Errorf("unable to get client: %v", err)
			}
			rootCmd.Println(client.String())
		} else {
			clients, err := ClientRepository.GetAll()
			if err != nil {
				return fmt.Errorf("unable to get clients: %v", err)
			}
			for _, client := range clients {
				rootCmd.Println(client.String())
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
	Args:    cobra.MaximumNArgs(1),
	Aliases: []string{"sessions"},
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := preRunCheckOutput()
		if err != nil {
			return err
		}

		reportTimeframe, err = ExtractTimeframeFromFlags()
		if err != nil {
			return err
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var sessions []models.Session
		var err error

		if len(args) == 1 {
			sessions = GetRelativeSessionsFromArgs(args, clientName)
		} else {
			sessions = GetSessionsWithTimeframe(*reportTimeframe)
		}
		if err != nil {
			return err
		}

		filteredSessions := FilterSessionsByClient(&sessions, clientName)
		SortSessions(filteredSessions, descendingOrder)
		content, err := generateView(filteredSessions)
		if err != nil {
			return err
		}
		rootCmd.Print(*content)
		return nil
	},
}

func generateView(slice *[]models.Session) (*string, error) {
	if len(*slice) == 0 {
		return nil, ErrNoAvailableData
	}
	var (
		buffer  *bytes.Buffer
		err     error
		content = ""
	)

	switch output {
	case "text":
		if summary {
			content, err = generateSummaryView(slice)
		} else {
			content, err = generateFullGetView(slice)
		}
	case "csv":
		if verbose {
			buffer, err = models.SerializeSessionsToFullCSV(*slice)
		} else {
			buffer, err = models.SerializeSessionsToCSV(*slice)
		}
		content = buffer.String()
	}
	if err != nil {
		return nil, err
	}
	return &content, nil
}

func generateSummaryView(slice *[]models.Session) (string, error) {
	if len(*slice) == 0 {
		return "", errors.New("no sessions to summarize")
	}

	buffer := new(bytes.Buffer)
	w := tabwriter.NewWriter(buffer, 0, 0, 1, ' ', tabwriter.TabIndent)
	if !hideHeaders {
		_, err := fmt.Fprintln(w, "DATE\tCLIENT\tTIME\tAMOUNT\tCURRENCY")
		if err != nil {
			return "", err
		}
	}

	clientData := make(map[string]struct {
		totalTime   time.Duration
		totalAmount float64
		lastDate    time.Time
		currency    string
	})
	currencyData := make(map[string]struct {
		totalTime   time.Duration
		totalAmount float64
	})

	for _, session := range *slice {
		client := session.Client.Name
		currency := session.Client.Currency

		if _, exists := clientData[client]; !exists {
			clientData[client] = struct {
				totalTime   time.Duration
				totalAmount float64
				lastDate    time.Time
				currency    string
			}{currency: currency}
		}
		if _, exists := currencyData[currency]; !exists {
			currencyData[currency] = struct {
				totalTime   time.Duration
				totalAmount float64
			}{}
		}

		data := clientData[client]
		sessionEnd := time.Now()
		if session.Finished() {
			sessionEnd = session.End
		}
		duration := sessionEnd.Sub(session.Start)
		data.totalTime += duration.Truncate(time.Second)
		currencyData[currency] = struct {
			totalTime   time.Duration
			totalAmount float64
		}{
			totalTime:   currencyData[currency].totalTime + duration,
			totalAmount: currencyData[currency].totalAmount,
		}
		earnings, err := session.Earnings()
		if err == nil {
			data.totalAmount += earnings
			currencyData[currency] = struct {
				totalTime   time.Duration
				totalAmount float64
			}{
				totalTime:   currencyData[currency].totalTime,
				totalAmount: currencyData[currency].totalAmount + earnings,
			}
		}
		if session.Start.After(data.lastDate) {
			data.lastDate = session.Start
		}
		clientData[client] = data
	}

	for client, data := range clientData {
		_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\t%s\n",
			data.lastDate.Format("2006-01-02"),
			client,
			data.totalTime,
			data.totalAmount,
			data.currency)
		if err != nil {
			return "", err
		}
	}
	if len(clientData) > 1 {
		for currency, data := range currencyData {
			_, err := fmt.Fprintf(w, "-\t<all>\t%s\t%.2f\t%s\n",
				data.totalTime,
				data.totalAmount,
				currency)
			if err != nil {
				return "", err
			}
		}
	}

	err := w.Flush()
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func generateFullGetView(slice *[]models.Session) (string, error) {
	buffer := new(bytes.Buffer)
	w := tabwriter.NewWriter(buffer, 0, 0, 1, ' ', tabwriter.TabIndent)
	if !hideHeaders {
		if verbose {
			_, err := fmt.Fprintln(w, "ID\tDATE\tCLIENT\tSTART\tEND\tDURATION\tAMOUNT\tCURRENCY\tNOTE")
			if err != nil {
				return "", err
			}
		} else {
			_, err := fmt.Fprintln(w, "DATE\tCLIENT\tDURATION\tAMOUNT\tCURRENCY")
			if err != nil {
				return "", err
			}
		}
	}
	for _, session := range *slice {
		id := fmt.Sprintf("%d", session.ID)

		earnings, _ := session.Earnings()
		endTime := "N/A"
		if session.Finished() {
			endTime = session.End.Format("15:04:05")
		}

		if verbose {
			_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%.2f\t%s\t%s\n",
				id,
				session.Start.Format("2006-01-02"),
				session.Client.Name,
				session.Start.Format("15:04:05"),
				endTime,
				session.Duration(),
				earnings,
				session.Client.Currency,
				session.Note,
			)
			if err != nil {
				return "", err
			}
		} else {
			_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%.2f\t%s\n",
				session.Start.Format("2006-01-02"),
				session.Client.Name,
				session.Duration(),
				earnings,
				session.Client.Currency,
			)
			if err != nil {
				return "", err
			}
		}
	}
	err := w.Flush()
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
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
	getSessionCmd.Flags().StringVarP(&clientName, "client", "c", "", "Specify the client name")
	getSessionCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	getSessionCmd.Flags().BoolVarP(&summary, "summary", "s", false, "Output summary of sessions")
	getSessionCmd.Flags().BoolVar(&hideHeaders, "hide-headers", false, "Output summary of sessions")
	getSessionCmd.Flags().BoolVar(&dayReport, "day", false, "Hide headers in report")
	getSessionCmd.Flags().BoolVar(&weekReport, "week", false, "Get report for this current week")
	getSessionCmd.Flags().StringVar(&monthReport, "month", "", "Get report for a specific month (format: YYYY-MM), leave empty for current month")
	getSessionCmd.Flags().StringVar(&yearReport, "year", "", "Get report for a specific year (format: YYYY), leave empty for current year")
	getSessionCmd.Flags().BoolVar(&allReport, "all", false, "Get all sessions")
	getSessionCmd.Flags().BoolVar(&descendingOrder, "desc", false, "Sort sessions in descending order (defaults to ascending order)")
	getSessionCmd.Flags().StringVarP(&output, "output", "o", "text", "Specify the output format")
	getSessionCmd.Flags().Lookup("month").NoOptDefVal = strconv.Itoa(int(currentMonth))
	getSessionCmd.Flags().Lookup("year").NoOptDefVal = strconv.Itoa(currentYear)
}
