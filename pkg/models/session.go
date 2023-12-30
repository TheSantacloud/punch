package models

import (
	"fmt"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Session struct {
	ID     *uint32
	Client Client
	Start  *time.Time
	End    *time.Time
	Note   string
}

func (s Session) Matches(session Session) bool {
	if session.ID == nil || s.ID == nil {
		return false
	}
	return *session.ID == *s.ID || s.Similar(session)
}

func (s Session) String() string {
	return fmt.Sprintf("Client: %s, Date: %s, Duration: %s",
		s.Client.Name,
		s.Start.Format("02/01/2006"),
		s.Duration(),
	)
}

func (s Session) Similar(session Session) bool {
	return s.Start.Format("02/01/2006 15:04:05") ==
		session.Start.Format("02/01/2006 15:04:05") &&
		s.Client.Name == session.Client.Name
}

func (s Session) Equals(session Session) bool {
	return (session.ID != nil && s.ID != nil && *session.ID == *s.ID) &&
		*s.Start == *session.Start &&
		s.End != nil && session.End != nil && *s.End == *session.End &&
		s.Client.Name == session.Client.Name &&
		s.Note == session.Note
}

func (s Session) Conflicts(session Session) bool {
	return (session.ID != nil && s.ID != nil && *session.ID == *s.ID) &&
		(((session.End != nil && s.End != nil) && (*s.End != *session.End)) ||
			s.Client.Name != session.Client.Name ||
			*s.Start != *session.Start)
}

func (s Session) Finished() bool {
	return s.End != nil
}

func (s Session) Summary() string {
	earnings := "Earnings: N/A"
	duration := "Duration: N/A"

	id := "<nil>"
	if s.ID != nil {
		id = strconv.Itoa(int(*s.ID))
	}

	value, err := s.Earnings()
	if err == nil {
		duration = s.Duration()
		earnings = fmt.Sprintf("%.2f %s", value, s.Client.Currency)
	}

	date := s.Start.Format("2006-01-02")
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s", id, date, s.Client.Name, duration, earnings)
}

func (s Session) Earnings() (float64, error) {
	if s.Start == nil || s.End == nil {
		return 0, fmt.Errorf("Session not started or ended")
	}
	delta := s.End.Sub(*s.Start)
	hours := delta.Hours()
	value := float64(s.Client.PPH) * hours
	return value, nil
}

func (s Session) Duration() string {
	if s.Start == nil || s.End == nil {
		return "N/A"
	}

	delta := s.End.Sub(*s.Start)
	hours := int(delta.Hours())
	minutes := int(delta.Minutes()) % 60
	seconds := int(delta.Seconds()) % 60

	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func (s Session) SerializeYAML() (*[]byte, error) {
	id := ""
	if s.ID != nil {
		id = fmt.Sprint(*s.ID)
	}
	end := ""
	if s.End != nil {
		end = s.End.Format("15:04:05")
	}
	ed := EditableSession{
		ID:        id,
		Client:    s.Client.Name,
		Date:      s.Start.Format("2006-01-02"),
		StartTime: s.Start.Format("15:04:05"),
		EndTime:   end,
		Note:      s.Note,
	}

	data, err := yaml.Marshal(ed)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
