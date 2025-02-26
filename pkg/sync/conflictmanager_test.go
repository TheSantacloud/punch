package sync

import (
	"testing"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestConflictManager_GetConflicts(t *testing.T) {
	localSessions := []models.Session{
		{Start: time.Now().Add(-2 * time.Hour)},
		{Start: time.Now()},
	}
	remoteSessions := []models.Session{
		{Start: time.Now().Add(-1 * time.Hour)},
		{Start: time.Now().Add(-3 * time.Hour)},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
}

func TestConflictManager_DetectDeletedSessions(t *testing.T) {
	existingSessions := &[]models.Session{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}
	editedSessions := &[]models.Session{
		{ID: 1},
		{ID: 3},
	}

	deletedSessions := DetectDeletedSessions(existingSessions, editedSessions)
	assert.Len(t, deletedSessions, 1)
	assert.Equal(t, uint32(2), deletedSessions[0].ID)
}

func TestConflictManager_GetConflicts_EmptySessions(t *testing.T) {
	localSessions := []models.Session{}
	remoteSessions := []models.Session{}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotContains(t, buf.String(), "<<<", "Buffer should not contain the word 'Conflict'")
}

func TestConflictManager_GetConflicts_IdenticalSessions(t *testing.T) {
	timeNow := time.Now()
	sessions := []models.Session{{Start: timeNow}}

	buf, err := GetConflicts(sessions, sessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotContains(t, buf.String(), "<<<", "Buffer should not contain the word 'Conflict'")
}

func TestConflictManager_DetectDeletedSessions_NoDeletions(t *testing.T) {
	sessions := &[]models.Session{
		{ID: 1},
		{ID: 2},
	}

	deletedSessions := DetectDeletedSessions(sessions, sessions)
	assert.Empty(t, deletedSessions, "There should be no deleted sessions")
}

func TestConflictManager_DetectDeletedSessions_AllDeletions(t *testing.T) {
	existingSessions := &[]models.Session{
		{ID: 1},
		{ID: 2},
	}
	editedSessions := &[]models.Session{}

	deletedSessions := DetectDeletedSessions(existingSessions, editedSessions)
	assert.Len(t, deletedSessions, 2, "All sessions should be detected as deleted")
}

func TestConflictManager_GetConflicts_NonOverlapping(t *testing.T) {
	localSessions := []models.Session{
		{ID: 1, Start: time.Now().Add(-2 * time.Hour), End: time.Now().Add(-1 * time.Hour)},
	}
	remoteSessions := []models.Session{
		{ID: 1, Start: time.Now().Add(-4 * time.Hour), End: time.Now().Add(-3 * time.Hour)},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for non-overlapping sessions")
}

func TestConflictManager_DetectDeletedSessions_MixedDeletions(t *testing.T) {
	existingSessions := &[]models.Session{
		{ID: 1},
		{ID: 2},
		{ID: 3},
	}
	editedSessions := &[]models.Session{
		{ID: 1},
		{ID: 3},
		{ID: 4}, // New session added
	}

	deletedSessions := DetectDeletedSessions(existingSessions, editedSessions)
	assert.Len(t, deletedSessions, 1, "One session should be detected as deleted")
	assert.Equal(t, uint32(2), deletedSessions[0].ID)
}

func TestConflictManager_GetConflicts_SameIDDifferentEndShowsMismatch(t *testing.T) {
	now := time.Now()
	localSessions := []models.Session{
		{ID: 1, Start: now, End: now.Add(1 * time.Hour)},
	}
	remoteSessions := []models.Session{
		{ID: 1, Start: now, End: now.Add(2 * time.Hour)},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for same ID with different end times")
}

func TestConflictManager_GetConflicts_SameIDDifferentStartShowsMismatch(t *testing.T) {
	now := time.Now()
	localSessions := []models.Session{
		{ID: 1, Start: now, End: now.Add(1 * time.Hour)},
	}
	remoteSessions := []models.Session{
		{ID: 1, Start: now.Add(-1 * time.Hour), End: now},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for same ID with different start times")
}

func TestConflictManager_GetConflicts_SameIDDiffererentCompanyMismatch(t *testing.T) {
	now := time.Now()
	localSessions := []models.Session{
		{ID: 1, Start: now, Client: models.Client{Name: "CompanyA"}},
	}
	remoteSessions := []models.Session{
		{ID: 1, Start: now, Client: models.Client{Name: "CompanyB"}},
	}

	buf, err := GetConflicts(localSessions, remoteSessions)
	assert.NoError(t, err)
	assert.NotNil(t, buf)
	assert.NotEmpty(t, buf.String(), "Buffer should show conflict for same ID but different companies")
}
