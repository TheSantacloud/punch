package sync

import (
	"github.com/dormunis/punch/pkg/models"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

func uint32Ptr(i uint32) *uint32 {
	return &i
}

func TestConflictManager_GetConflicts(t *testing.T) {
	localSessions := []models.Session{
		{Start: timePtr(time.Now().Add(-2 * time.Hour))},
		{Start: timePtr(time.Now())},
	}
	remoteSessions := []models.Session{
		{Start: timePtr(time.Now().Add(-1 * time.Hour))},
		{Start: timePtr(time.Now().Add(-3 * time.Hour))},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestConflictManager_DetectDeletedSessions(t *testing.T) {
	existingSessions := &[]models.Session{
		{ID: uint32Ptr(1)},
		{ID: uint32Ptr(2)},
		{ID: uint32Ptr(3)},
	}
	editedSessions := &[]models.Session{
		{ID: uint32Ptr(1)},
		{ID: uint32Ptr(3)},
	}

	deletedSessions := DetectDeletedSessions(existingSessions, editedSessions)
	assert.Len(t, deletedSessions, 1)
	assert.Equal(t, uint32(2), *deletedSessions[0].ID)
}

func TestConflictManager_GetConflicts_EmptySessions(t *testing.T) {
	localSessions := []models.Session{}
	remoteSessions := []models.Session{}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Empty(t, buf.String(), "Buffer should be empty for no conflicts")
}

func TestConflictManager_GetConflicts_IdenticalSessions(t *testing.T) {
	timeNow := time.Now()
	sessions := []models.Session{{Start: &timeNow}}

	buf, err := GetConflicts(sessions, sessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.Empty(t, buf.String(), "Buffer should be empty for identical sessions")
}

func TestConflictManager_DetectDeletedSessions_NoDeletions(t *testing.T) {
	sessions := &[]models.Session{
		{ID: uint32Ptr(1)},
		{ID: uint32Ptr(2)},
	}

	deletedSessions := DetectDeletedSessions(sessions, sessions)
	assert.Empty(t, deletedSessions, "There should be no deleted sessions")
}

func TestConflictManager_DetectDeletedSessions_AllDeletions(t *testing.T) {
	existingSessions := &[]models.Session{
		{ID: uint32Ptr(1)},
		{ID: uint32Ptr(2)},
	}
	editedSessions := &[]models.Session{}

	deletedSessions := DetectDeletedSessions(existingSessions, editedSessions)
	assert.Len(t, deletedSessions, 2, "All sessions should be detected as deleted")
}

func TestConflictManager_GetConflicts_NonOverlapping(t *testing.T) {
	localSessions := []models.Session{
		{Start: timePtr(time.Now().Add(-2 * time.Hour)), End: timePtr(time.Now().Add(-1 * time.Hour))},
	}
	remoteSessions := []models.Session{
		{Start: timePtr(time.Now().Add(-4 * time.Hour)), End: timePtr(time.Now().Add(-3 * time.Hour))},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for non-overlapping sessions")
}

func TestConflictManager_DetectDeletedSessions_MixedDeletions(t *testing.T) {
	existingSessions := &[]models.Session{
		{ID: uint32Ptr(1)},
		{ID: uint32Ptr(2)},
		{ID: uint32Ptr(3)},
	}
	editedSessions := &[]models.Session{
		{ID: uint32Ptr(1)},
		{ID: uint32Ptr(3)},
		{ID: uint32Ptr(4)}, // New session added
	}

	deletedSessions := DetectDeletedSessions(existingSessions, editedSessions)
	assert.Len(t, deletedSessions, 1, "One session should be detected as deleted")
	assert.Equal(t, uint32(2), *deletedSessions[0].ID)
}

func TestConflictManager_GetConflicts_SameIDDifferentEndShowsMismatch(t *testing.T) {
	now := time.Now()
	localSessions := []models.Session{
		{ID: uint32Ptr(1), Start: timePtr(now), End: timePtr(now.Add(1 * time.Hour))},
	}
	remoteSessions := []models.Session{
		{ID: uint32Ptr(1), Start: timePtr(now), End: timePtr(now.Add(2 * time.Hour))},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for same ID with different end times")
}

func TestConflictManager_GetConflicts_SameIDDifferentStartShowsMismatch(t *testing.T) {
	now := time.Now()
	localSessions := []models.Session{
		{ID: uint32Ptr(1), Start: timePtr(now), End: timePtr(now.Add(1 * time.Hour))},
	}
	remoteSessions := []models.Session{
		{ID: uint32Ptr(1), Start: timePtr(now.Add(-1 * time.Hour)), End: timePtr(now)},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for same ID with different start times")
}

func TestConflictManager_GetConflicts_DifferentIDSameStartAndClientShowsMismatch(t *testing.T) {
	now := time.Now()
	clientName := "ClientA"
	localSessions := []models.Session{
		{ID: uint32Ptr(1), Start: timePtr(now), Client: models.Client{Name: clientName}},
	}
	remoteSessions := []models.Session{
		{ID: uint32Ptr(2), Start: timePtr(now), Client: models.Client{Name: clientName}},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for different IDs with same start time and client")
}

func TestConflictManager_GetConflicts_LocalIDButNoRemoteIDAndSimilarTimesMismatch(t *testing.T) {
	now := time.Now()
	localSessions := []models.Session{
		{ID: uint32Ptr(1), Start: timePtr(now)},
	}
	remoteSessions := []models.Session{
		{Start: timePtr(now.Add(5 * time.Minute))},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for local ID with no remote ID and similar times")
}

func TestConflictManager_GetConflicts_SameIDDiffererentCompanyMismatch(t *testing.T) {
	now := time.Now()
	localSessions := []models.Session{
		{ID: uint32Ptr(1), Start: timePtr(now), Client: models.Client{Name: "CompanyA"}},
	}
	remoteSessions := []models.Session{
		{ID: uint32Ptr(1), Start: timePtr(now), Client: models.Client{Name: "CompanyB"}},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for same ID but different companies")
}
