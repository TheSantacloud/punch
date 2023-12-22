package models

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Session struct {
	ID      uint32
	Company Company
	Start   *time.Time
	End     *time.Time
	Note    string
}

type EditableSession struct {
	ID        string `yaml:"id"`
	Company   string `yaml:"company"`
	Date      string `yaml:"date"`
	StartTime string `yaml:"start_time"`
	EndTime   string `yaml:"end_time"`
	Note      string `yaml:"note"`
}

func (d Session) Summary() string {
	earnings := "Earnings: N/A"
	duration := "Duration: N/A"

	value, err := d.Earnings()
	if err == nil {
		duration = d.Duration()
		earnings = fmt.Sprintf("%.2f %s", value, d.Company.Currency)
	}

	date := d.Start.Format("2006-01-02")
	return fmt.Sprintf("%s\t%s\t%s\t%s", date, d.Company.Name, duration, earnings)
}

func (d Session) Earnings() (float64, error) {
	if d.Start == nil || d.End == nil {
		return 0, fmt.Errorf("Session not started or ended")
	}
	delta := d.End.Sub(*d.Start)
	hours := delta.Hours()
	value := float64(d.Company.PPH) * hours
	return value, nil
}

func (d Session) Duration() string {
	if d.Start == nil || d.End == nil {
		return "N/A"
	}

	delta := d.End.Sub(*d.Start)
	hours := int(delta.Hours())
	minutes := int(delta.Minutes()) % 60
	seconds := int(delta.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func SerializeSessionsToYAML(sessions []Session) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("# Change either the `start_time` or `end_time` fields to edit the day\n")
	buf.WriteString("# The `id`, `company` and `date` fields are for reference only\n")
	buf.WriteString("\n")
	for i, session := range sessions {
		ed := EditableSession{
			ID:        fmt.Sprint(session.ID),
			Company:   session.Company.Name,
			Date:      session.Start.Format("2006-01-02"),
			StartTime: session.Start.Format("15:04:05"),
			EndTime:   session.End.Format("15:04:05"),
			Note:      session.Note,
		}

		data, err := yaml.Marshal(ed)
		if err != nil {
			return nil, err
		}

		buf.Write(data)
		if i < len(sessions)-1 {
			buf.WriteString("---\n")
		}
	}

	return &buf, nil
}

func DeserializeSessionsFromYAML(buf *bytes.Buffer, sessions *[]Session) error {
	decoder := yaml.NewDecoder(buf)

	for {
		var ed EditableSession
		err := decoder.Decode(&ed)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		fmt.Printf("%s\n", ed)
		if ed.ID == "" {
			return fmt.Errorf("missing ID field for session")
		}

		id, err := strconv.ParseUint(ed.ID, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid ID format for session: %s", ed.ID)
		}

		updated := false
		for i, session := range *sessions {
			if session.ID == uint32(id) {
				startTime, err := time.Parse("15:04:05 2006-01-02", ed.StartTime+" "+ed.Date)
				if err != nil {
					return err
				}
				endTime, err := time.Parse("15:04:05 2006-01-02", ed.EndTime+" "+ed.Date)
				if err != nil {
					return err
				}

				if startTime.After(endTime) || startTime.Equal(endTime) {
					return fmt.Errorf("start time must be before end time")
				}

				(*sessions)[i].Start = &startTime
				(*sessions)[i].End = &endTime
				(*sessions)[i].Note = ed.Note
				updated = true
				break
			}
		}

		if !updated {
			return fmt.Errorf("could not find entry for session ID '%d'", id)
		}
	}

	return nil
}

func SerializeSessionsToCSV(sessions []Session) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("company,date,duration\n")
	for _, session := range sessions {
		buf.WriteString(fmt.Sprintf("%s,%s,%s\n",
			session.Company.Name,
			session.Start.Format("2006-01-02"),
			session.Duration(),
		))
	}

	return &buf, nil
}

func SerializeSessionsToFullCSV(session []Session) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("company,date,start_time,end_time,hours,earnings,currency,note\n")
	for _, session := range session {
		earnings, _ := session.Earnings()

		buf.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%.2f,%s,%s\n",
			session.Company.Name,
			session.Start.Format("2006-01-02"),
			session.Start.Format("15:04:05"),
			session.End.Format("15:04:05"),
			session.Duration(),
			earnings,
			session.Company.Currency,
			session.Note,
		))
	}

	return &buf, nil
}
