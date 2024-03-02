package cli

import (
	"testing"
	"time"

	"github.com/dormunis/punch/pkg/models"
	"github.com/dormunis/punch/pkg/repositories"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func timePtr(t time.Time) *time.Time {
	return &t
}

func calculateDelta(start, end time.Time) time.Duration {
	delta := end.Sub(start)
	delta = delta - (delta % time.Second) // truncate milliseconds
	return delta
}

func TestTimeQuery_ExtractTime_CountDelta_InvalidHeadInput_Invalid(t *testing.T) {
	client := models.Client{Name: "test"}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)
	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
		GetLastSessions(gomock.Any(), gomock.Any()).
		Times(0)

	_, _, err := ExtractTime("-x", &client)
	assert.Error(t, err)
}

func TestTimeQuery_ExtractTime_CountDelta_InvalidHeadInput_Empty(t *testing.T) {
	client := models.Client{Name: "test"}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)
	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
		GetLastSessions(gomock.Any(), gomock.Any()).
		Times(0)

	_, _, err := ExtractTime("-", &client)
	assert.Error(t, err)
}

func TestTimeQuery_ExtractTime_CountDelta_2(t *testing.T) {
	client := models.Client{Name: "test"}
	sessions := []models.Session{
		{Start: timePtr(time.Now().Add(-2 * time.Hour))},
		{Start: timePtr(time.Now())},
	}
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	count := uint32(2)

	SessionRepository = repositories.NewMockSessionRepository(mockCtrl)
	SessionRepository.(*repositories.MockSessionRepository).EXPECT().
		GetLastSessions(count, &client).
		Return(&sessions, nil)

	start, end, err := ExtractTime("-2", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)
}

func TestTimeQuery_ExtractTime_TimeDelta_InvalidInput(t *testing.T) {
	client := models.Client{Name: "test"}

	_, _, err := ExtractTime("-2x", &client)
	assert.Error(t, err)
}

func TestTimeQuery_ExtractTime_TimeDelta_2_seconds(t *testing.T) {
	client := models.Client{Name: "test"}

	start, end, err := ExtractTime("-2s", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)

	delta := calculateDelta(start, end)
	assert.Equal(t, 2*time.Second, delta)
}

func TestTimeQuery_ExtractTime_TimeDelta_2_minutes(t *testing.T) {
	client := models.Client{Name: "test"}

	start, end, err := ExtractTime("-2m", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)

	delta := calculateDelta(start, end)
	assert.Equal(t, 2*time.Minute, delta)
}

func TestTimeQuery_ExtractTime_TimeDelta_2_hours(t *testing.T) {
	client := models.Client{Name: "test"}

	start, end, err := ExtractTime("-2h", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)

	delta := calculateDelta(start, end)
	assert.Equal(t, 2*time.Hour, delta)
}

func TestTimeQuery_ExtractTime_TimeDelta_2_days(t *testing.T) {
	client := models.Client{Name: "test"}

	start, end, err := ExtractTime("-2d", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)

	delta := calculateDelta(start, end)
	assert.Equal(t, 24*2*time.Hour, delta)
}

func TestTimeQuery_ExtractTime_TimeDelta_2_weeks(t *testing.T) {
	client := models.Client{Name: "test"}

	start, end, err := ExtractTime("-2w", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)

	delta := calculateDelta(start, end)
	assert.Equal(t, 24*7*2*time.Hour, delta)
}

func TestTimeQuery_ExtractTime_TimeDelta_2_months(t *testing.T) {
	client := models.Client{Name: "test"}

	start, end, err := ExtractTime("-2M", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)

	delta := calculateDelta(start, end)
	assert.Equal(t, 24*30*2*time.Hour, delta)
}

func TestTimeQuery_ExtractTime_TimeDelta_2_years(t *testing.T) {
	client := models.Client{Name: "test"}

	start, end, err := ExtractTime("-2y", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)

	delta := calculateDelta(start, end)
	assert.Equal(t, 24*365*2*time.Hour, delta)
}

func TestTimeQuery_ExtractTime_TimeDelta_2_5_years(t *testing.T) {
	client := models.Client{Name: "test"}

	start, end, err := ExtractTime("-2.5y", &client)
	assert.NoError(t, err)
	assert.NotNil(t, start)
	assert.NotNil(t, end)

	delta := calculateDelta(start, end)
	assert.Equal(t, 24*365*2.5*time.Hour, delta)
}

func TestTimeQuery_ExtractTime_InvalidTime(t *testing.T) {
	client := models.Client{Name: "test"}
	_, _, err := ExtractTime("2", &client)
	assert.Error(t, err)
}
