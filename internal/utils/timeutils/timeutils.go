package timeutils

import "time"

func GetFirstDayOfMonth() time.Time {
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	return time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
}

func ParseDate(date string, fallback time.Time) time.Time {
	parsedDate, parseErr := time.Parse(time.RFC3339, date)
	if parseErr != nil {
		return fallback
	}
	return parsedDate
}
