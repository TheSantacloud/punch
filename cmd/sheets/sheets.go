package sheets

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dormunis/consulting/cmd/daily"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	Service       *sheets.Service
	SpreadsheetId string
	SheetName     string
}

const (
	DATE_COLUMN  = "C"
	START_COLUMN = "D"
	END_COLUMN   = "E"
	PPH          = 500
)

func GetSheet() (*Sheet, error) {
	ctx := context.Background()

	client, err := getClient(ctx)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		os.Exit(1)
	}
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		return nil, err
	}

	sheet := Sheet{
		Service:       srv,
		SpreadsheetId: "1mkK5xy5YN_Jp8P_lACic6ZT4bGvSrbY6TsVp3Vlg1CQ",
		SheetName:     "test",
	}
	return &sheet, nil
}

func getClient(ctx context.Context) (*http.Client, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to find home directory: %v", err)
	}
	path := filepath.Join(homeDir, ".work", "service-account.json")
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Unable to read service account key file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to parse service account key file to config: %v", err)
	}
	return config.Client(ctx), nil
}

func (s *Sheet) GetTodaysRow() (*daily.Today, error) {
	rangeString := s.SheetName + "!" + DATE_COLUMN + ":" + END_COLUMN
	resp, err := s.Service.Spreadsheets.Values.Get(s.SpreadsheetId, rangeString).Do()
	if err != nil {
		err = fmt.Errorf("Unable to retrieve data from sheet: %v", err)
		return nil, err
	}

	if len(resp.Values) == 0 {
		err = fmt.Errorf("Malformed sheet")
		return nil, err
	}

	todaysDate := time.Now().Format("02/01/2006")
	for i, row := range resp.Values {
		if row[0] == todaysDate {
			end := ""
			start := ""
			if len(row) == 3 {
				end = row[2].(string)
				start = row[1].(string)
			} else if len(row) == 2 {
				start = row[1].(string)
			} else {
				err = fmt.Errorf("Malformed sheet")
			}

			today := daily.Today{
				Row:   i,
				Date:  row[0].(string),
				Start: start,
				End:   end,
			}
			return &today, nil
		}
	}

	today := daily.Today{
		Row:   len(resp.Values),
		Date:  todaysDate,
		Start: "",
		End:   "",
	}

	return &today, nil
}

func (s *Sheet) UpdateRow(today *daily.Today) error {
	rowStr := strconv.Itoa(today.Row + 1)

	vr := &sheets.ValueRange{
		Values: [][]interface{}{
			{today.Date, today.Start, today.End},
		},
	}

	updateRange := s.SheetName + "!" + DATE_COLUMN + rowStr + ":" + END_COLUMN + rowStr

	_, err := s.Service.Spreadsheets.Values.Update(
		s.SpreadsheetId, updateRange, vr).
		ValueInputOption("USER_ENTERED").Do()

	if err != nil {
		return err
	}

	return nil
}
