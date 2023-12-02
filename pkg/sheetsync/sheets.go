package sheetsync

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/database"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	Service       *sheets.Service
	SpreadsheetId string
	SheetName     string

	Columns struct {
		Company   string
		Date      string
		StartTime string
		EndTime   string
		TotalTime string
	}
}

type Record struct {
	Day database.Day
	Row int
}

func (r Record) Matches(day database.Day) bool {
	return r.Day.Start.Format("02/01/2006") == day.Start.Format("02/01/2006") &&
		r.Day.Company.Name == day.Company.Name
}

var (
	companyColumnIndex   int
	dateColumnIndex      int
	startTimeColumnIndex int
	endTimeColumnIndex   int
	totalTimeColumnIndex int
)

func GetSheet(cfg config.SpreadsheetSettings) (*Sheet, error) {
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
		SpreadsheetId: cfg.ID,
		SheetName:     cfg.Sheet,
		Columns: struct {
			Company   string
			Date      string
			StartTime string
			EndTime   string
			TotalTime string
		}{
			Company:   cfg.Columns.Company,
			Date:      cfg.Columns.Date,
			StartTime: cfg.Columns.StartTime,
			EndTime:   cfg.Columns.EndTime,
			TotalTime: cfg.Columns.TotalTime,
		},
	}
	return &sheet, nil
}

func (s *Sheet) UpsertDay(day database.Day, records *[]Record) error {
	for _, record := range *records {
		if record.Day.Start == nil {
			continue
		}
		if record.Matches(day) {
			return s.updateDay(day, record)
		}
	}
	return s.insertDay(day)
}

func (s *Sheet) updateDay(day database.Day, record Record) error {
	if day.Start == nil || day.End == nil {
		return fmt.Errorf("Day must be complete")
	}
	var vr sheets.ValueRange
	rangeToUpdate := fmt.Sprintf("%s!A%d", s.SheetName, record.Row+1)
	vr.Values = append(vr.Values, s.dayToRow(day))
	_, err := s.Service.Spreadsheets.Values.Update(s.SpreadsheetId,
		rangeToUpdate, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}
	return nil
}

func (s *Sheet) insertDay(day database.Day) error {
	if day.Start == nil || day.End == nil {
		return fmt.Errorf("Day must be complete")
	}
	var vr sheets.ValueRange
	vr.Values = append(vr.Values, s.dayToRow(day))
	_, err := s.Service.Spreadsheets.Values.Append(s.SpreadsheetId,
		s.SheetName, &vr).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		return err
	}
	return nil
}

func (s *Sheet) dayToRow(day database.Day) []interface{} {
	maxIdx := max(companyColumnIndex, dateColumnIndex, startTimeColumnIndex,
		endTimeColumnIndex, totalTimeColumnIndex) + 1
	row := make([]interface{}, maxIdx)
	for i := range row {
		switch i {
		case companyColumnIndex:
			row[companyColumnIndex] = day.Company.Name
		case dateColumnIndex:
			row[dateColumnIndex] = day.Start.Format("02/01/2006")
		case startTimeColumnIndex:
			row[startTimeColumnIndex] = day.Start.Format("15:04:05")
		case endTimeColumnIndex:
			row[endTimeColumnIndex] = day.End.Format("15:04:05")
		case totalTimeColumnIndex:
			row[totalTimeColumnIndex] = day.Duration()
		default:
			row[i] = ""
		}
	}
	return row
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

func (s *Sheet) ParseSheet(records *[]Record) error {
	resp, err := s.readSheet()
	if err != nil {
		return err
	}

	for i, row := range resp.Values {
		if i == 0 {
			s.parseHeaders(row)
			continue
		}

		if len(row) < 3 {
			continue
		}

		day, err := s.parseDayFromRow(row)
		if err != nil {
			return err
		}
		record := Record{Day: *day, Row: i}
		*records = append(*records, record)
	}
	return nil
}

func (s *Sheet) parseHeaders(row []interface{}) {
	for i, column := range row {
		switch column {
		case s.Columns.Company:
			companyColumnIndex = i
		case s.Columns.Date:
			dateColumnIndex = i
		case s.Columns.StartTime:
			startTimeColumnIndex = i
		case s.Columns.EndTime:
			endTimeColumnIndex = i
		case s.Columns.TotalTime:
			totalTimeColumnIndex = i
		}
	}
}

func (s *Sheet) parseDayFromRow(row []interface{}) (*database.Day, error) {
	if len(row) < 3 {
		return nil, fmt.Errorf("Invalid row")
	}

	var startTime time.Time
	startTimestamp := row[startTimeColumnIndex].(string) + " " + row[dateColumnIndex].(string)
	startTime, err := time.Parse("15:04:05 02/01/2006", startTimestamp)
	if err != nil {
		startTime, err = time.Parse("02/01/2006", row[dateColumnIndex].(string))
		if err != nil {
			return nil, err
		}
	}

	var endTime time.Time
	endTimestamp := row[endTimeColumnIndex].(string) + " " + row[dateColumnIndex].(string)
	endTime, _ = time.Parse("15:04:05 02/01/2006", endTimestamp)

	day := database.Day{
		Company: database.Company{Name: row[companyColumnIndex].(string)},
		Start:   &startTime,
		End:     &endTime,
	}
	return &day, nil
}

func (s *Sheet) readSheet() (*sheets.ValueRange, error) {
	resp, err := s.Service.Spreadsheets.Values.Get(s.SpreadsheetId, s.SheetName).Do()
	if err != nil {
		return nil, err
	}
	if len(resp.Values) == 0 {
		return nil, fmt.Errorf("No data found")
	}
	return resp, nil
}
