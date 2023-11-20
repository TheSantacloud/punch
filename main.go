package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	Service       *sheets.Service
	SpreadsheetId string
	SheetName     string
}

type Today struct {
	Row int
	// TODO: add company
	Date  string
	Start string
	End   string
}

const (
	DATE_COLUMN  = "C"
	START_COLUMN = "D"
	END_COLUMN   = "E"
	// TODO: make this grab from the spreadsheet as well/configurable
	PPH = 500
)

func main() {
	ctx := context.Background()

	client, err := getClient(ctx)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		os.Exit(1)
	}

	sheet, err := getSheet(ctx, client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		os.Exit(1)
	}

	today, err := sheet.getTodaysRow()
	if err != nil {
		log.Fatalf("%v", err)
		os.Exit(1)
	}

	// TODO: add comment with cli
	today.Update()
	sheet.UpdateRow(today)

	if today.End != "" {
		printSummary(today)
	}

}

func getClient(ctx context.Context) (*http.Client, error) {
	// TODO: make this configurable
	b, err := os.ReadFile("service-account.json")
	if err != nil {
		log.Fatalf("Unable to read service account key file: %v", err)
	}

	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to parse service account key file to config: %v", err)
	}
	return config.Client(ctx), nil
}

func getSheet(ctx context.Context, client *http.Client) (*Sheet, error) {
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
		return nil, err
	}

	// TODO: add configurable company in col A

	sheet := Sheet{
		Service: srv,
		// TODO: make these configurable
		SpreadsheetId: "1mkK5xy5YN_Jp8P_lACic6ZT4bGvSrbY6TsVp3Vlg1CQ",
		SheetName:     "hours",
	}
	return &sheet, nil
}

func (s *Sheet) getTodaysRow() (*Today, error) {
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
			} else if len(row) == 2 {
				start = row[1].(string)
			} else {
				err = fmt.Errorf("Malformed sheet")
			}

			today := Today{
				Row:   i,
				Date:  row[0].(string),
				Start: start,
				End:   end,
			}
			return &today, nil
		}
	}

	today := Today{
		Row:   len(resp.Values),
		Date:  todaysDate,
		Start: "",
		End:   "",
	}

	return &today, nil
}

func (t *Today) Update() {
	if t.Start == "" {
		t.Start = time.Now().Format("15:04:05")
		fmt.Println("Checked in at", t.Start)
		return
	}
	if t.End == "" {
		t.End = time.Now().Format("15:04:05")
		fmt.Println("Checked out at", t.End)
		return
	}
}

func (s *Sheet) UpdateRow(today *Today) error {
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

func printSummary(today *Today) {
	start, _ := time.Parse("15:04:05", today.Start)
	end, _ := time.Parse("15:04:05", today.End)
	duration := end.Sub(start)
	moneyMade := fmt.Sprintf("( %.2f ₪ )", duration.Hours()*PPH)
	fmt.Println("Worked for", duration, moneyMade)
}
