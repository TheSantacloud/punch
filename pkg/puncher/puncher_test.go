package puncher

import (
	"testing"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPuncher_ToggleCheckInOut_NoRunningSessionStartsOne(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)
	client := models.Client{Name: "Test"}

	mockRepo.EXPECT(). // toggler looks for the latest session
				GetLatestSession().
				Return(nil, repositories.ErrSessionNotFound).
				Times(1)

	mockRepo.EXPECT(). // start looks for all sessions within the provided timestamp
				GetLatestSessionOnSpecificDate(gomock.Any(), gomock.Eq(client)).
				Return(nil, repositories.ErrSessionNotFound).
				Times(1)

	mockRepo.EXPECT().
		Insert(gomock.Any(), false).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Times(0)

	session, err := puncher.ToggleCheckInOut(&client, "Testing Toggle")

	assert.NoError(t, err, "ToggleCheckInOut should not return an error")
	assert.NotNil(t, session, "Session should not be nil")
	assert.NotNil(t, session.Start, "Session start time should not be nil")
	assert.Nil(t, session.End, "Session end time should be nil")
	assert.Equal(t, client.Name, session.Client.Name, "Client name should match")
}

func TestPuncher_ToggleCheckInOut_RunningSessionFinalizesIt(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)
	client := models.Client{Name: "Test"}

	startTime := time.Now()
	runningSession := models.Session{
		Client: client,
		Start:  &startTime,
	}

	mockRepo.EXPECT().
		GetLatestSession().
		Return(&runningSession, nil).
		Times(1)

	mockRepo.EXPECT().
		Insert(gomock.Any(), gomock.Any()).
		Times(0)

	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Times(1)

	session, err := puncher.ToggleCheckInOut(&client, "Testing Toggle")

	assert.NoError(t, err, "ToggleCheckInOut should not return an error")
	assert.NotNil(t, session, "Session should not be nil")
	assert.NotNil(t, session.Start, "Session start time should not be nil")
	assert.NotNil(t, session.End, "Session end time should not be nil")
	assert.Equal(t, client.Name, session.Client.Name, "Client name should match")
}

func TestPuncher_ToggleCheckInOut_AlreadyFinishedSessionCreatesANewOne(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)
	client := models.Client{Name: "Test"}

	startTime := time.Now().Add(-time.Hour)
	endTime := time.Now()
	previousSession := models.Session{
		Client: client,
		Start:  &startTime,
		End:    &endTime,
	}

	mockRepo.EXPECT().
		GetLatestSession().
		Return(&previousSession, nil).
		Times(1)

	mockRepo.EXPECT().
		GetLatestSessionOnSpecificDate(gomock.Any(), gomock.Eq(client)).
		Return(&previousSession, nil).
		Times(1)

	mockRepo.EXPECT().
		Insert(gomock.Any(), false).
		Return(nil).
		Times(1)

	mockRepo.EXPECT().
		Update(gomock.Any(), gomock.Any()).
		Times(0)

	session, err := puncher.ToggleCheckInOut(&client, "Testing Toggle")

	assert.NoError(t, err, "ToggleCheckInOut should not return an error")
	assert.NotNil(t, session, "Session should not be nil")
	assert.NotNil(t, session.Start, "Session start time should not be nil")
	assert.Nil(t, session.End, "Session end time should be nil")
	assert.Equal(t, client.Name, session.Client.Name, "Client name should match")
	assert.Less(t, *previousSession.Start, *session.Start, "Previous session start time is not less than new session start time")
}

func TestPuncher_StartSession_NoPreviousSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)
	client := models.Client{Name: "Test"}
	now := time.Now()

	mockRepo.EXPECT().
		GetLatestSessionOnSpecificDate(gomock.Any(), gomock.Eq(client)).
		Return(nil, repositories.ErrSessionNotFound).
		Times(1)

	mockRepo.EXPECT().
		Insert(gomock.Any(), false).
		Return(nil).
		Times(1)

	session, err := puncher.StartSession(client, now, "")

	assert.NoError(t, err, "StartSession should not return an error")
	assert.NotNil(t, session, "Session should not be nil")
	assert.Equal(t, client.Name, session.Client.Name, "Client name should match")
	assert.Equal(t, now, *session.Start, "Session start time should match")
}

func TestPuncher_StartSession_PreviousSessionNotEndedYet(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)
	client := models.Client{Name: "Test"}
	now := time.Now()

	startTime := now.Add(-time.Hour)
	previousSession := models.Session{
		Client: client,
		Start:  &startTime,
	}

	mockRepo.EXPECT().
		GetLatestSessionOnSpecificDate(gomock.Any(), gomock.Eq(client)).
		Return(&previousSession, nil).
		Times(1)

	session, err := puncher.StartSession(client, now, "")

	assert.Nil(t, session, "Session should be nil")
	assert.Error(t, err, "StartSession should return an error")
	assert.Equal(t, ErrSessionAlreadyStarted, err, "StartSession should return ErrSessionAlreadyStarted")
}

func TestPuncher_StartSession_PreviousSessionInTimespaceAlreadyEndedStartsANewOne(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)
	client := models.Client{Name: "Test"}
	now := time.Now()

	startTime := now.Add(-time.Hour)
	endTime := now.Add(-time.Minute)
	previousSession := models.Session{
		Client: client,
		Start:  &startTime,
		End:    &endTime,
	}

	mockRepo.EXPECT().
		GetLatestSessionOnSpecificDate(gomock.Any(), gomock.Eq(client)).
		Return(&previousSession, nil).
		Times(1)

	mockRepo.EXPECT().
		Insert(gomock.Any(), false).
		Return(nil).
		Times(1)

	session, err := puncher.StartSession(client, now, "")

	assert.NoError(t, err, "StartSession should not return an error")
	assert.NotNil(t, session, "Session should not be nil")
	assert.Equal(t, client.Name, session.Client.Name, "Client name should match")
	assert.NotEqual(t, *previousSession.Start, *session.Start, "Previous session start time should not match new session start time")
}

func TestPuncher_EndSession_FinalizedGoodSession(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)

	client := models.Client{Name: "Test"}
	now := time.Now()
	startTime := now.Add(-time.Hour)
	previousSession := models.Session{
		Client: client,
		Start:  &startTime,
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), false).
		Return(nil).
		Times(1)

	session, err := puncher.EndSession(previousSession, now, "")

	assert.NoError(t, err, "StartSession should not return an error")
	assert.NotNil(t, session, "Session should not be nil")
	assert.Equal(t, client.Name, session.Client.Name, "Client name should match")
	assert.Equal(t, *previousSession.Start, *session.Start, "Session start time should match previous session start time")
	assert.Less(t, *session.Start, *session.End, "Session start time should be less than session end time")
}

func TestPuncher_EndSession_GettingAnAlreadyEndedSessionDoesNothing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)

	client := models.Client{Name: "Test"}
	now := time.Now()
	startTime := now.Add(-time.Hour)
	endTime := now.Add(-time.Minute)
	previousSession := models.Session{
		Client: client,
		Start:  &startTime,
		End:    &endTime,
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), false).
		Return(nil).
		Times(0)

	mockRepo.EXPECT().
		Insert(gomock.Any(), false).
		Return(nil).
		Times(0)

	session, err := puncher.EndSession(previousSession, now, "")

	assert.Error(t, err, "EndSession should return an error")
	assert.Nil(t, session, "Session should be nil")
	assert.Equal(t, ErrSessionAlreadyEnded, err, "EndSession should return ErrSessionAlreadyEnded")
}

func TestPuncher_EndSession_SessionStartTimeIsAfterEndTimeDoesNothing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockRepo := repositories.NewMockSessionRepository(mockCtrl)
	puncher := NewPuncher(mockRepo)

	client := models.Client{Name: "Test"}
	endTime := time.Now().Add(-time.Hour)
	startTime := time.Now().Add(-time.Minute)
	previousSession := models.Session{
		Client: client,
		Start:  &startTime,
	}

	mockRepo.EXPECT().
		Update(gomock.Any(), false).
		Return(nil).
		Times(0)

	mockRepo.EXPECT().
		Insert(gomock.Any(), false).
		Return(nil).
		Times(0)

	session, err := puncher.EndSession(previousSession, endTime, "")

	assert.Error(t, err, "EndSession should return an error")
	assert.Nil(t, session, "Session should be nil")
	assert.Equal(t, ErrInvalidSession, err, "EndSession should return ErrInvalidSession")
}
