package sheets

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/dormunis/punch/pkg/config"
	"github.com/dormunis/punch/pkg/models"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Columns struct {
	ID        string
	Client    string
	Date      string
	StartTime string
	EndTime   string
	TotalTime string
	Note      string
}

type Sheet struct {
	Service       *sheets.Service
	SpreadsheetId string
	SheetName     string
	Columns       Columns
}

type Record struct {
	Session models.Session
	Row     int
}

var (
	idColumnIndex        int
	clientColumnIndex    int
	dateColumnIndex      int
	startTimeColumnIndex int
	endTimeColumnIndex   int
	totalTimeColumnIndex int
	noteColumnIndex      int
)

func NewSheet(cfg config.SpreadsheetRemote) (*Sheet, error) {
	srv, err := CreateGoogleSheetClient(cfg.ServiceAccountJsonPath)
	if err != nil {
		return nil, err
	}

	sheet, err := GetSheet(srv, cfg)
	if err != nil {
		return nil, err
	}

	return sheet, nil
}

func CreateGoogleSheetClient(serviceAccountJsonPath string) (*sheets.Service, error) {
	ctx := context.Background()
	client, err := getClient(ctx, serviceAccountJsonPath)
	if err != nil {
		return nil, err
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func getClient(ctx context.Context, serviceAccountJsonPath string) (*http.Client, error) {
	b, err := os.ReadFile(serviceAccountJsonPath)
	if err != nil {
		log.Fatalf("Unable to read service account key file: %v", err)
	}
	config, err := google.JWTConfigFromJSON(b, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatalf("Unable to parse service account key file to config: %v", err)
	}
	return config.Client(ctx), nil
}

func GetSheet(srv *sheets.Service, cfg config.SpreadsheetRemote) (*Sheet, error) {
	sheet := Sheet{
		Service:       srv,
		SpreadsheetId: cfg.ID,
		SheetName:     cfg.SheetName,
		Columns: Columns{
			Client:    cfg.Columns.Client,
			Date:      cfg.Columns.Date,
			StartTime: cfg.Columns.StartTime,
			EndTime:   cfg.Columns.EndTime,
			TotalTime: cfg.Columns.TotalTime,
			Note:      cfg.Columns.Note,
		},
	}
	return &sheet, nil
}

func (s *Sheet) ParseSheet(records *[]Record) error {
	resp, err := s.readSheet()
	if err != nil {
		return err
	}

	for i, row := range resp.Values {
		if i == 0 {
			s.ParseHeaders(row)
			continue
		}

		if len(row) < 4 {
			continue
		}

		session, err := s.SessionFromRow(row)
		if err != nil {
			return err
		}
		record := Record{Session: *session, Row: i}
		*records = append(*records, record)
	}
	return nil
}

func (s *Sheet) SessionToRow(session models.Session) []interface{} {
	maxIdx := max(clientColumnIndex, dateColumnIndex, startTimeColumnIndex,
		endTimeColumnIndex, totalTimeColumnIndex, noteColumnIndex) + 1
	row := make([]interface{}, maxIdx)
	for i := range row {
		switch i {
		case idColumnIndex:
			row[idColumnIndex] = strconv.FormatUint(uint64(session.ID), 10)
		case clientColumnIndex:
			row[clientColumnIndex] = session.Client.Name
		case dateColumnIndex:
			row[dateColumnIndex] = session.Start.Format("02/01/2006")
		case startTimeColumnIndex:
			row[startTimeColumnIndex] = session.Start.Format("15:04:05")
		case endTimeColumnIndex:
			if !session.Finished() {
				row[endTimeColumnIndex] = ""
			} else {
				row[endTimeColumnIndex] = session.End.Format("15:04:05")
			}
		case totalTimeColumnIndex:
			if !session.Finished() {
				row[totalTimeColumnIndex] = ""
			} else {
				row[totalTimeColumnIndex] = session.Duration()
			}
		case noteColumnIndex:
			row[noteColumnIndex] = session.Note
		default:
			row[i] = ""
		}
	}
	return row
}

func (s *Sheet) SessionFromRow(row []interface{}) (*models.Session, error) {
	if len(row) < 4 {
		return nil, fmt.Errorf("Invalid row")
	}

	var id uint32
	if row[idColumnIndex] != "" {
		value, err := strconv.ParseUint(row[idColumnIndex].(string), 10, 32)
		if err != nil {
			return nil, err
		}
		id = uint32(value)
	}

	var startTime time.Time
	startTimestamp := row[startTimeColumnIndex].(string) + " " + row[dateColumnIndex].(string)
	startTime, err := time.ParseInLocation("15:04:05 02/01/2006", startTimestamp, time.Local)
	if err != nil {
		startTime, err = time.ParseInLocation("02/01/2006", row[dateColumnIndex].(string), time.Local)
		if err != nil {
			return nil, err
		}
	}

	var endTime time.Time
	if len(row) > 4 && row[endTimeColumnIndex] != "" {
		endTimestamp := row[endTimeColumnIndex].(string) + " " + row[dateColumnIndex].(string)
		parsedTime, err := time.ParseInLocation("15:04:05 02/01/2006", endTimestamp, time.Local)
		if err != nil {
			return nil, err
		}
		endTime = parsedTime
	}

	if endTime != models.NULL_TIME && endTime.Before(startTime) {
		endTime = endTime.AddDate(0, 0, 1)
	}

	var note string
	if len(row) > noteColumnIndex {
		note = row[noteColumnIndex].(string)
	} else {
		note = ""
	}

	session := models.Session{
		ID:     id,
		Client: models.Client{Name: row[clientColumnIndex].(string)},
		Start:  startTime,
		End:    endTime,
		Note:   note,
	}
	return &session, nil
}

func (s *Sheet) AddRow(session models.Session) error {
	// TODO: keep sheet sorted
	row := s.SessionToRow(session)
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{row},
	}

	_, err := s.Service.Spreadsheets.Values.Append(s.SpreadsheetId, s.SheetName, valueRange).ValueInputOption("USER_ENTERED").Do()
	return err
}

func (s *Sheet) UpdateRow(record Record) error {
	row := s.SessionToRow(record.Session)
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{row},
	}

	// Adding one because of the header row
	rangeToUpdate := fmt.Sprintf("%s!A%d", s.SheetName, record.Row+1)
	_, err := s.Service.Spreadsheets.Values.Update(s.SpreadsheetId, rangeToUpdate, valueRange).ValueInputOption("USER_ENTERED").Do()
	return err
}

func (s *Sheet) ParseHeaders(row []interface{}) {
	for i, column := range row {
		switch column {
		case s.Columns.ID:
			idColumnIndex = i
		case s.Columns.Client:
			clientColumnIndex = i
		case s.Columns.Date:
			dateColumnIndex = i
		case s.Columns.StartTime:
			startTimeColumnIndex = i
		case s.Columns.EndTime:
			endTimeColumnIndex = i
		case s.Columns.TotalTime:
			totalTimeColumnIndex = i
		case s.Columns.Note:
			noteColumnIndex = i
		}
	}
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
