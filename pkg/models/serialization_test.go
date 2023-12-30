package models

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEditableSession_ToSession_Valid(t *testing.T) {
	ed := EditableSession{
		ID:        "1",
		Client:    "Test Client",
		Date:      "2022-01-02",
		StartTime: "09:00:00",
		EndTime:   "17:00:00",
		Note:      "Test note",
	}
	session, err := ed.ToSession()
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, uint32(1), *session.ID)
}

func TestEditableSession_ToSession_InvalidID(t *testing.T) {
	ed := EditableSession{ID: "invalid"}
	_, err := ed.ToSession()
	assert.Error(t, err)
}

func TestEditableSession_ToSession_InvalidTimeFormat(t *testing.T) {
	ed := EditableSession{
		ID:        "1",
		Client:    "Test Client",
		Date:      "2022-01-02",
		StartTime: "invalid",
		EndTime:   "17:00:00",
	}
	_, err := ed.ToSession()
	assert.Error(t, err)
}

func TestEditableSession_ToSession_StartAfterEnd(t *testing.T) {
	ed := EditableSession{
		ID:        "1",
		Client:    "Test Client",
		Date:      "2022-01-02",
		StartTime: "17:00:00",
		EndTime:   "09:00:00",
	}
	_, err := ed.ToSession()
	assert.Error(t, err)
}

func TestSerializeSessionsToYAML_ValidSessions(t *testing.T) {
	sessions := []Session{sampleSession(), sampleSession()}
	buf, err := SerializeSessionsToYAML(sessions)

	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Contains(t, buf.String(), sessions[0].Client.Name)
}

func TestDeserializeSessionsFromYAML_ValidYAML(t *testing.T) {
	sessions := []Session{sampleSession(), sampleSession()}
	buf, _ := SerializeSessionsToYAML(sessions)
	deserializedSessions, err := DeserializeSessionsFromYAML(buf)

	assert.NoError(t, err)
	assert.Len(t, *deserializedSessions, 2)
}

func TestDeserializeSessionsFromYAML_InvalidYAML(t *testing.T) {
	buf := bytes.NewBufferString("invalid yaml content")
	_, err := DeserializeSessionsFromYAML(buf)

	assert.Error(t, err)
}

func TestDeserializeAndUpdateSessionsFromYAML_ValidData(t *testing.T) {
	sessions := []Session{sampleSession(), sampleSession()}
	buf, _ := SerializeSessionsToYAML(sessions)
	err := DeserializeAndUpdateSessionsFromYAML(buf, &sessions)

	assert.NoError(t, err)
}

func TestDeserializeAndUpdateSessionsFromYAML_MissingID(t *testing.T) {
	buf := bytes.NewBufferString("client: Test Client\ndate: 2022-01-02\nstart_time: 09:00:00\nend_time: 17:00:00\n")
	sessions := []Session{sampleSession()}
	err := DeserializeAndUpdateSessionsFromYAML(buf, &sessions)

	assert.Error(t, err)
}

func TestSerializeSessionsToCSV_ValidSessions(t *testing.T) {
	sessions := []Session{sampleSession(), sampleSession()}
	buf, err := SerializeSessionsToCSV(sessions)

	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Contains(t, buf.String(), sessions[0].Client.Name)
}

func TestSerializeSessionsToFullCSV_ValidSessions(t *testing.T) {
	sessions := []Session{sampleSession(), sampleSession()}
	buf, err := SerializeSessionsToFullCSV(sessions)

	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Contains(t, buf.String(), sessions[0].Client.Name)
}

func TestDeserializeSessionsFromYAML_EmptyYAML(t *testing.T) {
	buf := bytes.NewBufferString("")
	deserializedSessions, err := DeserializeSessionsFromYAML(buf)

	assert.NoError(t, err)
	assert.Empty(t, *deserializedSessions, "Deserialized sessions should be empty for empty YAML")
}

func TestDeserializeAndUpdateSessionsFromYAML_InvalidIDFormat(t *testing.T) {
	buf := bytes.NewBufferString("id: invalid\nclient: Test Client\ndate: 2022-01-02\nstart_time: 09:00:00\nend_time: 17:00:00\n")
	sessions := []Session{sampleSession()}
	err := DeserializeAndUpdateSessionsFromYAML(buf, &sessions)

	assert.Error(t, err, "Should error out for invalid ID format")
}

func TestDeserializeAndUpdateSessionsFromYAML_NonExistentSessionID(t *testing.T) {
	buf := bytes.NewBufferString("id: 999\nclient: Test Client\ndate: 2022-01-02\nstart_time: 09:00:00\nend_time: 17:00:00\n")
	sessions := []Session{sampleSession()}
	err := DeserializeAndUpdateSessionsFromYAML(buf, &sessions)

	assert.Error(t, err, "Should error out for non-existent session ID")
}

func TestSerializeSessionsToCSV_EmptySessions(t *testing.T) {
	sessions := []Session{}
	buf, err := SerializeSessionsToCSV(sessions)

	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Contains(t, buf.String(), "client,date,duration\n", "CSV should contain headers even for empty session list")
}

func TestSerializeSessionsToFullCSV_EmptySessions(t *testing.T) {
	sessions := []Session{}
	buf, err := SerializeSessionsToFullCSV(sessions)

	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Contains(t, buf.String(), "client,date,start_time,end_time,hours,earnings,currency,note\n", "CSV should contain headers even for empty session list")
}

func TestSerializeSessionsToYAML_EmptySessions(t *testing.T) {
	sessions := []Session{}
	buf, err := SerializeSessionsToYAML(sessions)

	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Contains(t, buf.String(), "# Change either the `start_time` or `end_time` fields to edit the day\n", "YAML should contain headers even for empty session list")
}
