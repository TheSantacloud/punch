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
	Client    string `yaml:"client"`
	Date      string `yaml:"date"`
	StartTime string `yaml:"start_time"`
	EndTime   string `yaml:"end_time"`
	Note      string `yaml:"note"`
}

func (ed EditableSession) ToSession() (*Session, error) {
	var uintId uint32
	if ed.ID != "" {
		id, err := strconv.ParseUint(ed.ID, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("invalid ID format for session: %s", ed.ID)
		}
		uintId = uint32(id)
	}

	client := Client{Name: ed.Client}
	startTime, err := time.ParseInLocation("15:04:05 2006-01-02", ed.StartTime+" "+ed.Date, time.Local)
	if err != nil {
		return nil, err
	}
	var endTime time.Time
	if ed.EndTime != "N/A" {
		endTime, err = time.ParseInLocation("15:04:05 2006-01-02", ed.EndTime+" "+ed.Date, time.Local)
	}
	if err != nil {
		return nil, err
	}

	if endTime != NULL_TIME && (startTime.After(endTime) || startTime.Equal(endTime)) {
		nextDayEndTime := endTime.AddDate(0, 0, 1)
		endTime = nextDayEndTime
	}

	return &Session{
		ID:     uintId,
		Client: client,
		Start:  startTime,
		End:    endTime,
		Note:   ed.Note,
	}, nil
}

func SerializeSessionsToYAML(sessions []Session) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("# Change either the `start_time` or `end_time` fields to edit the day\n")
	buf.WriteString("# The `id`, `client` and `date` fields are for reference only\n")
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
		if err != nil {
			return nil, err
		}
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
			if session.ID != uint32(id) {
				continue
			}
			err = updateSession(ed, &session)
			if err != nil {
				return err
			}
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
		endTime = endTime.AddDate(0, 0, 1)
	}

	session.Start = startTime
	session.End = endTime
	session.Note = ed.Note
	return nil
}

func SerializeSessionsToCSV(sessions []Session) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("date,client,duration,amount,currency\n")
	for _, session := range sessions {
		earnings, err := session.Earnings()
		if err != nil {
			return nil, err
		}
		buf.WriteString(fmt.Sprintf("%s,%s,%s,%.2f,%s\n",
			session.Client.Name,
			session.Start.Format("2006-01-02"),
			session.Duration(),
			earnings,
			session.Client.Currency,
		))
	}

	return &buf, nil
}

func SerializeSessionsToFullCSV(session []Session) (*bytes.Buffer, error) {
	var buf bytes.Buffer

	buf.WriteString("id,date,client,start_time,end_time,duration,amount,currency,note\n")
	for _, session := range session {
		id := fmt.Sprintf("%d", session.ID)

		earnings, err := session.Earnings()
		earningsString := "N/A"
		if err == nil {
			earningsString = fmt.Sprintf("%.2f", earnings)
		}

		end := "N/A"
		if session.End != NULL_TIME {
			end = session.End.Format("15:04:05")
		}

		buf.WriteString(fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			id,
			session.Start.Format("2006-01-02"),
			session.Client.Name,
			session.Start.Format("15:04:05"),
			end,
			session.Duration(),
			earningsString,
			session.Client.Currency,
			session.Note,
		))
	}

	return &buf, nil
}
