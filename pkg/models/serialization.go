package models

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	YAML_SERIALIZATION_SEPARATOR = "---\n"
)

type EditableSession struct {
	ID        string `yaml:"id"`
	Company   string `yaml:"company"`
	Date      string `yaml:"date"`
	StartTime string `yaml:"start_time"`
	EndTime   string `yaml:"end_time"`
	Note      string `yaml:"note"`
}

func (ed EditableSession) ToSession() (*Session, error) {
	id, err := strconv.ParseUint(ed.ID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format for session: %s", ed.ID)
	}
	uintId := uint32(id)

	company := Company{Name: ed.Company}
	startTime, err := time.Parse("15:04:05 2006-01-02", ed.StartTime+" "+ed.Date)
	if err != nil {
		return nil, err
	}
	endTime, err := time.Parse("15:04:05 2006-01-02", ed.EndTime+" "+ed.Date)
	if err != nil {
		return nil, err
	}

	if startTime.After(endTime) || startTime.Equal(endTime) {
		return nil, fmt.Errorf("start time must be before end time")
	}

	return &Session{
		ID:      &uintId,
		Company: company,
		Start:   &startTime,
		End:     &endTime,
		Note:    ed.Note,
	}, nil
}

func SerializeSessionsToYAML(sessions []Session) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("# Change either the `start_time` or `end_time` fields to edit the day\n")
	buf.WriteString("# The `id`, `company` and `date` fields are for reference only\n")
	buf.WriteString("\n")
	for i, session := range sessions {
		serialized, err := session.SerializeYAML()
		if err != nil {
			return nil, err
		}
		buf.Write(*serialized)
		if i < len(sessions)-1 {
			buf.WriteString(YAML_SERIALIZATION_SEPARATOR)
		}
	}

	return &buf, nil
}

func DeserializeSessionsFromYAML(buf *bytes.Buffer) (*[]Session, error) {
	decoder := yaml.NewDecoder(buf)
	var sessions []Session
	sessions = make([]Session, 0)

	for {
		var ed EditableSession
		err := decoder.Decode(&ed)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		session, err := ed.ToSession()
		sessions = append(sessions, *session)
	}
	return &sessions, nil
}

func DeserializeAndUpdateSessionsFromYAML(buf *bytes.Buffer, sessions *[]Session) error {
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

		if ed.ID == "" {
			return fmt.Errorf("missing ID field for session")
		}

		id, err := strconv.ParseUint(ed.ID, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid ID format for session: %s", ed.ID)
		}

		updated := false
		for _, session := range *sessions {
			if *session.ID != uint32(id) {
				continue
			}
			updateSession(ed, &session)
			updated = true
			break
		}

		if !updated {
			return fmt.Errorf("could not find entry for session ID '%d'", id)
		}
	}

	return nil
}

func updateSession(ed EditableSession, session *Session) error {
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

	session.Start = &startTime
	session.End = &endTime
	session.Note = ed.Note
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
