package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func executeCommand(t *testing.T, args []string) (string, error) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs(args)
	err := rootCmd.Execute()
	if err != nil {
		return buf.String(), err
	}
	return buf.String(), nil
}

func createSampleSession() models.Session {
	today := time.Now()
	now := time.Date(today.Year(), today.Month(), today.Day(), 9, 0, 0, 0, time.UTC)
	later := now.Add(8 * time.Hour)
	client := models.Client{Name: "Test Client", Currency: "USD", PPH: 42069}
	id := uint32(1)

	return models.Session{
		ID:     id,
		Client: client,
		Start:  now,
		End:    later,
		Note:   "Test Note",
	}
}

func TestCli_GetSession_NoPreexistingSessionsReturnsNothing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)

	returnValue := []models.Session{}

	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
		GetAllSessionsBetweenDates(gomock.Any(), gomock.Any()).
		Return(&returnValue, nil).
		Times(1)

	args := []string{"get", "session"}
	_, err := executeCommand(t, args)
	assert.ErrorIs(t, err, NoAvailableDataError)
}

func TestCli_GetSession_GetSessionWithoutASessionForThatDayReturnsNothing(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)

	unclosedSession := createSampleSession()
	newStart := unclosedSession.Start.AddDate(0, 0, -1)
	newEnd := unclosedSession.End.AddDate(0, 0, -1)
	unclosedSession.Start = newStart
	unclosedSession.End = newEnd

	returnValue := []models.Session{
		unclosedSession,
	}

	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
		GetAllSessionsBetweenDates(gomock.Any(), gomock.Any()).
		Return(&returnValue, nil).
		Times(1)

	args := []string{"get", "session"}
	_, err := executeCommand(t, args)
	assert.ErrorIs(t, err, NoAvailableDataError)
}

func TestCli_GetSession_DayFlagQueriesForDayOnly(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)

	today := time.Now()
	beginningOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	returnVal := []models.Session{}

	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
		GetAllSessionsBetweenDates(beginningOfDay, gomock.Any()).
		Return(&returnVal, nil).
		Times(1)

	args := []string{"get", "session", "--day"}
	_, err := executeCommand(t, args)
	assert.ErrorIs(t, err, NoAvailableDataError)
}

// func TestCli_GetSession_WeekFlagQueriesForWeekOnly(t *testing.T) {
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()
// 	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)
//
// 	today := time.Now()
// 	beginningOfWeek := time.Date(today.Year(), today.Month(), today.Day()-int(today.Weekday()), 0, 0, 0, 0, today.Location())
// 	returnVal := []models.Session{}
//
// 	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
// 		GetAllSessionsBetweenDates(beginningOfWeek, gomock.Any()).
// 		Return(&returnVal, nil).
// 		Times(1)
//
// 	args := []string{"get", "session", "--week"}
// 	_, err := executeCommand(t, args)
// 	assert.ErrorIs(t, err, NoAvailableDataError)
// }
//
// func TestCli_GetSession_MonthFlagQueriesForMonthOnly(t *testing.T) {
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()
// 	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)
//
// 	today := time.Now()
// 	beginningOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
// 	returnVal := []models.Session{}
//
// 	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
// 		GetAllSessionsBetweenDates(beginningOfMonth, gomock.Any()).
// 		Return(&returnVal, nil).
// 		Times(1)
//
// 	args := []string{"get", "session", "--month"}
// 	_, err := executeCommand(t, args)
// 	assert.ErrorIs(t, err, NoAvailableDataError)
// }
//
// func TestCli_GetSession_YearFlagQueriesForYearOnly(t *testing.T) {
// 	mockCtrl := gomock.NewController(t)
// 	defer mockCtrl.Finish()
// 	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)
//
// 	today := time.Now()
// 	beginningOfYear := time.Date(today.Year(), 1, 1, 0, 0, 0, 0, today.Location())
// 	returnVal := []models.Session{}
//
// 	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
// 		GetAllSessionsBetweenDates(beginningOfYear, gomock.Any()).
// 		Return(&returnVal, nil).
// 		Times(1)
//
// 	args := []string{"get", "session", "--year"}
// 	_, err := executeCommand(t, args)
// 	assert.ErrorIs(t, err, NoAvailableDataError)
// }
