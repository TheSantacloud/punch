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
}

type EditableDay struct {
	Company   string `yaml:"company"`
	Date      string `yaml:"date"`
	StartTime string `yaml:"start_time"`
	EndTime   string `yaml:"end_time"`
}

func (d Day) Summary() string {
	earnings := "Earnings: N/A"
	duration := "Duration: N/A"

	if d.Start != nil && d.End != nil {
		delta := d.End.Sub(*d.Start)
		hours := delta.Hours()
		duration = fmt.Sprintf("%.2f hours", hours)
		value := float64(d.Company.PPH) * hours
		earnings = fmt.Sprintf("₪%.2f", value)
	}

	return fmt.Sprintf("%s (%s)", duration, earnings)
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

func (d Day) FullSummary() string {
	summary := fmt.Sprintf("Company: %s\n", d.Company.Name)
	date := "Date: N/A"
	startTime := "Start: Not Started"
	endTime := "End: In Progress"
	duration := "Duration: N/A"
	earnings := "Earnings: N/A"

	if d.Start != nil {
		date = "Date: " + d.Start.Format("2006-01-02")
		startTime = "Start: " + d.Start.Format("15:04:05")

		if d.End != nil {
			endTime = "End: " + d.End.Format("15:04:05")
			delta := d.End.Sub(*d.Start)
			hours := delta.Hours()
			duration = fmt.Sprintf("Duration: %.2f hours", hours)
			value := float64(d.Company.PPH) * hours
			earnings = fmt.Sprintf("Earnings: ₪%.2f", value)
		}
	}

	return summary + fmt.Sprintf("%s\n%s\n%s\n%s\n%s", date, startTime, endTime, duration, earnings)
}

func SerializeDaysToYAML(dailies []Day) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("# Change either the `start_time` or `end_time` fields to edit the day\n")
	buf.WriteString("# The `company` and `date` fields are for reference only\n")
	buf.WriteString("\n")
	for _, day := range dailies {
		ed := struct {
			Company   string `yaml:"company"`
			Date      string `yaml:"date"`
			StartTime string `yaml:"start_time"`
			EndTime   string `yaml:"end_time"`
		}{
			Company:   day.Company.Name,
			Date:      day.Start.Format("2006-01-02"),
			StartTime: day.Start.Format("15:04:05"),
			EndTime:   day.End.Format("15:04:05"),
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

func DeserializeDaysFromYAML(buf *bytes.Buffer, dailies *[]Day) error {
	decoder := yaml.NewDecoder(buf)

	for {
		var ed struct {
			Company   string `yaml:"company"`
			Date      string `yaml:"date"`
			StartTime string `yaml:"start_time"`
			EndTime   string `yaml:"end_time"`
		}

		err := decoder.Decode(&ed)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		if ed.Company == "" && ed.Date == "" && ed.StartTime == "" && ed.EndTime == "" {
			continue
		}

		updated := false
		for i, day := range *dailies {
			if day.Company.Name == ed.Company && day.Start.Format("2006-01-02") == ed.Date {
				startTime, err := time.Parse("15:04:05 2006-01-02", ed.StartTime+" "+ed.Date)
				if err != nil {
					return err
				}
				endTime, err := time.Parse("15:04:05 2006-01-02", ed.EndTime+" "+ed.Date)
				if err != nil {
					return err
				}

				(*dailies)[i].Start = &startTime
				(*dailies)[i].End = &endTime
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
