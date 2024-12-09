package timeutils_test

import (
	"testing"
	"time"

	"github.com/basemind-ai/monorepo/shared/go/timeutils"
	"gotest.tools/v3/assert"
)

func TestGetFirstDayOfMonth(t *testing.T) {
	now := time.Now()
	year, month, _ := now.Date()
	location := now.Location()

	t.Run("get first day of the month", func(t *testing.T) {
		assert.Equal(
			t,
			time.Date(year, month, 1, 0, 0, 0, 0, location),
			timeutils.GetFirstDayOfMonth(),
		)
	})
}

func TestParseDate(t *testing.T) {
	testCases := []struct {
		fallback time.Time
		expected time.Time
		date     string
	}{
		{
			date:     "2022-01-01T00:00:00Z",
			expected: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
			fallback: time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
		{
			date:     "",
			expected: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
			fallback: time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, testCase := range testCases {
		t.Run("test date parser", func(t *testing.T) {
			assert.Equal(
				t,
				testCase.expected,
				timeutils.ParseDate(testCase.date, testCase.fallback),
			)
		})
	}
}
