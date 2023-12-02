package database

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

type Day struct {
	Company Company
	Start   *time.Time
	End     *time.Time
	Note    string
}

type EditableDay struct {
	Company   string `yaml:"company"`
	Date      string `yaml:"date"`
	StartTime string `yaml:"start_time"`
	EndTime   string `yaml:"end_time"`
	Note      string `yaml:"note"`
}

func (d Day) Summary() string {
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

func (d Day) Earnings() (float64, error) {
	if d.Start == nil || d.End == nil {
		return 0, fmt.Errorf("day not started or ended")
	}
	delta := d.End.Sub(*d.Start)
	hours := delta.Hours()
	value := float64(d.Company.PPH) * hours
	return value, nil
}

func (d Day) Duration() string {
	if d.Start == nil || d.End == nil {
		return "N/A"
	}

	delta := d.End.Sub(*d.Start)
	hours := int(delta.Hours())
	minutes := int(delta.Minutes()) % 60
	seconds := int(delta.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func SerializeDaysToYAML(days []Day) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("# Change either the `start_time` or `end_time` fields to edit the day\n")
	buf.WriteString("# The `company` and `date` fields are for reference only\n")
	buf.WriteString("\n")
	for _, day := range days {
		ed := struct {
			Company   string `yaml:"company"`
			Date      string `yaml:"date"`
			StartTime string `yaml:"start_time"`
			EndTime   string `yaml:"end_time"`
			Note      string `yaml:"note"`
		}{
			Company:   day.Company.Name,
			Date:      day.Start.Format("2006-01-02"),
			StartTime: day.Start.Format("15:04:05"),
			EndTime:   day.End.Format("15:04:05"),
			Note:      day.Note,
		}

		data, err := yaml.Marshal(ed)
		if err != nil {
			return nil, err
		}

		buf.Write(data)
		buf.WriteString("---\n")
	}

	return &buf, nil
}

func DeserializeDaysFromYAML(buf *bytes.Buffer, days *[]Day) error {
	decoder := yaml.NewDecoder(buf)

	for {
		var ed struct {
			Company   string `yaml:"company"`
			Date      string `yaml:"date"`
			StartTime string `yaml:"start_time"`
			EndTime   string `yaml:"end_time"`
			Note      string `yaml:"note"`
		}

		err := decoder.Decode(&ed)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if ed.Company == "" &&
			ed.Date == "" &&
			ed.StartTime == "" &&
			ed.EndTime == "" &&
			ed.Note == "" {
			continue
		}

		updated := false
		for i, day := range *days {
			if day.Company.Name == ed.Company && day.Start.Format("2006-01-02") == ed.Date {
				startTime, err := time.Parse("15:04:05 2006-01-02", ed.StartTime+" "+ed.Date)
				if err != nil {
					return err
				}
				endTime, err := time.Parse("15:04:05 2006-01-02", ed.EndTime+" "+ed.Date)
				if err != nil {
					return err
				}

				(*days)[i].Start = &startTime
				(*days)[i].End = &endTime
				(*days)[i].Note = ed.Note
				updated = true
				break
			}
		}

		if !updated {
			return fmt.Errorf("could not find entry for company '%s' on date '%s'", ed.Company, ed.Date)
		}
	}

	return nil
}

func SerializeDaysToCSV(days []Day) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("company,date,duration\n")
	for _, day := range days {
		buf.WriteString(fmt.Sprintf("%s,%s,%s\n",
			day.Company.Name,
			day.Start.Format("2006-01-02"),
			day.Duration(),
		))
	}

	return &buf, nil
}

func SerializeDaysToFullCSV(days []Day) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("company,date,start_time,end_time,hours,earnings,note\n")
	for _, day := range days {
		earnings, _ := day.Earnings()

		buf.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%.2f,%s\n",
			day.Company.Name,
			day.Start.Format("2006-01-02"),
			day.Start.Format("15:04:05"),
			day.End.Format("15:04:05"),
			day.Duration(),
			earnings,
			day.Note,
		))
	}

	return &buf, nil
}
