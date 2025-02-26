package models

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	NULL_TIME = time.Time{}
)

type Session struct {
	ID     uint32
	Client Client
	Start  time.Time
	End    time.Time
	Note   string
}

func (s Session) Matches(session Session) bool {
	return session.ID == s.ID || s.Similar(session)
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
	return session.ID == s.ID &&
		s.Start == session.Start &&
		s.End == session.End &&
		s.Client.Name == session.Client.Name &&
		s.Note == session.Note
}

func (s Session) Conflicts(session Session) bool {
	if s.ID != session.ID {
		return false
	}
	conflictReasons := []string{}
	if s.Start != NULL_TIME && s.Start != session.Start {
		conflictReasons = append(conflictReasons, "different start times")
	}
	if s.End != NULL_TIME && session.End != NULL_TIME && s.End != session.End {
		conflictReasons = append(conflictReasons, "different end times")
	}
	if s.Client.Name != session.Client.Name {
		conflictReasons = append(conflictReasons, "different client names")
	}
	if len(conflictReasons) > 0 {
		fmt.Printf("Session ID %d conflict: %s\n", s.ID, fmt.Sprintf("%v", conflictReasons))
		return true
	}
	return false
}

func (s Session) Finished() bool {
	return s.End != NULL_TIME
}

func (s Session) Earnings() (float64, error) {
	if s.Start == NULL_TIME {
		return 0, fmt.Errorf("Session not started or ended")
	}
	end := time.Now()
	if s.End != NULL_TIME {
		end = s.End
	}
	delta := end.Sub(s.Start)
	hours := delta.Hours()
	value := float64(s.Client.PPH) * hours
	return value, nil
}

func (s Session) Duration() string {
	if s.Start == NULL_TIME {
		return "N/A"
	}

	end := time.Now()
	if s.End != NULL_TIME {
		end = s.End
	}

	if end.Before(s.Start) {
		end = end.AddDate(0, 0, 1)
	}

	delta := end.Sub(s.Start)
	hours := int(delta.Hours())
	minutes := int(delta.Minutes()) % 60
	seconds := int(delta.Seconds()) % 60

	returnValue := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	if s.End == NULL_TIME {
		returnValue = "~" + returnValue
	}
	return returnValue
}

func (s Session) SerializeYAML() (*[]byte, error) {
	id := fmt.Sprint(s.ID)
	startDate := "N/A"
	startTime := "N/A"
	if s.Start != NULL_TIME {
		startDate = s.Start.Format("2006-01-02")
		startTime = s.Start.Format("15:04:05")
	}
	end := "N/A"
	if s.End != NULL_TIME {
		end = s.End.Format("15:04:05")
	}
	ed := EditableSession{
		ID:        id,
		Client:    s.Client.Name,
		Date:      startDate,
		StartTime: startTime,
		EndTime:   end,
		Note:      s.Note,
	}

	data, err := yaml.Marshal(ed)
	if err != nil {
		return nil, err
	}
	return &data, nil
}
