package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func sampleSession() Session {
	now := time.Date(2022, time.January, 1, 9, 0, 0, 0, time.UTC)
	later := now.Add(2 * time.Hour)
	client := Client{Name: "Test Client", Currency: "USD", PPH: 42069}
	id := uint32(1)
	return Session{
		ID:     id,
		Client: client,
		Start:  now,
		End:    later,
		Note:   "Test Note",
	}
}

func TestSession_Matches_ValidID(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()

	assert.True(t, session1.Matches(session2))
}

func TestSession_Similar_True(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()

	assert.True(t, session1.Similar(session2))
}

func TestSession_Equals_True(t *testing.T) {
	session := sampleSession()

	assert.True(t, session.Equals(session))
}

func TestSession_Equals_SameIDSameTime(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	session2.Start = session1.Start
	session2.End = session1.End

	assert.True(t, session1.Equals(session2))
}

func TestSession_Equals_SameIDDifferentStartTimes(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	startTime := session1.Start.Add(-time.Minute)
	session2.Start = startTime

	assert.False(t, session1.Equals(session2))
}

func TestSession_Equals_SameIDDifferentEndTimes(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	endTime := session1.End.Add(-time.Minute)
	session2.End = endTime

	assert.False(t, session1.Equals(session2))
}

func TestSession_Equals_FalseDifferentEnd(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	differentTime := session1.End.Add(time.Minute)
	session2.End = differentTime

	assert.False(t, session1.Equals(session2))
}

func TestSession_Conflicts_SameIDDifferentEnd(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	differentTime := session1.End.Add(time.Minute)
	session2.End = differentTime

	assert.True(t, session1.Conflicts(session2))
}

func TestSession_Conflicts_DifferentClientName(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	session2.Client.Name = "Different Client"

	assert.True(t, session1.Conflicts(session2))
}

func TestSession_Finished_True(t *testing.T) {
	session := sampleSession()
	assert.True(t, session.Finished(), "Session should be marked as finished")
}

func TestSession_Finished_False(t *testing.T) {
	session := sampleSession()
	session.End = NULL_TIME
	assert.False(t, session.Finished(), "Session should not be marked as finished")
}

func TestSession_Earnings_WithValidSession(t *testing.T) {
	session := sampleSession()
	earnings, err := session.Earnings()
	assert.NoError(t, err, "Earnings calculation should not produce an error")
	assert.Greater(t, earnings, 0.0, "Earnings should be greater than zero")
}

func TestSession_Duration_WithValidSession(t *testing.T) {
	session := sampleSession()
	duration := session.Duration()
	assert.NotEqual(t, duration, "N/A", "Duration should not be N/A")
}

func TestSession_Duration_WithNilStart(t *testing.T) {
	session := sampleSession()
	session.Start = NULL_TIME
	duration := session.Duration()
	assert.Equal(t, duration, "N/A", "Duration should be N/A")
}

func TestSession_SerializeYAML_ValidSession(t *testing.T) {
	session := sampleSession()
	data, err := session.SerializeYAML()
	assert.NoError(t, err, "Serialization should not produce an error")
	assert.NotNil(t, data, "Serialized data should not be nil")
}

func TestSession_SerializeYAML_WithNilStart(t *testing.T) {
	session := sampleSession()
	data, err := session.SerializeYAML()
	assert.NoError(t, err, "Serialization should produce an error")
	assert.NotNil(t, data, "Serialized data should not be nil")
}

func TestSession_Similar_DifferentClients(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	session2.Client.Name = "Different Client"

	assert.False(t, session1.Similar(session2), "Should return false if clients are different")
}

func TestSession_Equals_DifferentNotes(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	session2.Note = "Different Note"

	assert.False(t, session1.Equals(session2), "Should return false if notes are different")
}

func TestSession_Conflicts_SameTimes(t *testing.T) {
	session1 := sampleSession()
	session2 := sampleSession()
	session2.Start = session1.Start
	session2.End = session1.End

	assert.False(t, session1.Conflicts(session2), "Should return false if sessions have the same end time")
}

func TestSession_Earnings_WithNilStart(t *testing.T) {
	session := sampleSession()
	session.Start = NULL_TIME
	_, err := session.Earnings()

	assert.Error(t, err, "Earnings calculation should produce an error with nil Start")
}

func TestSession_SerializeYAML_WithNilEnd(t *testing.T) {
	session := sampleSession()
	session.End = NULL_TIME
	data, err := session.SerializeYAML()

	assert.NoError(t, err, "Serialization should not produce an error with nil End")
	assert.NotNil(t, data, "Serialized data should not be nil")
}
