package utils

import (
	"fmt"
	"time"
)

func IsTimeOngoing(startTime, endTime, date string) (bool, error) {
	const dateLayout = "2006-01-02"
	const timeLayout = "15:04"

	baseDate, err := time.Parse(dateLayout, date)
	if err != nil {
		return false, fmt.Errorf("invalid date format: %w", err)
	}

	start, err := time.Parse(timeLayout, startTime)
	if err != nil {
		return false, fmt.Errorf("invalid start time format: %w", err)
	}

	end, err := time.Parse(timeLayout, endTime)
	if err != nil {
		return false, fmt.Errorf("invalid end time format: %w", err)
	}

	startDateTime := time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), start.Hour(), start.Minute(), 0, 0, time.Local)
	var endDateTime time.Time

	if end.After(start) || end.Equal(start) {
		endDateTime = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day(), end.Hour(), end.Minute(), 0, 0, time.Local)
	} else {
		endDateTime = time.Date(baseDate.Year(), baseDate.Month(), baseDate.Day()+1, end.Hour(), end.Minute(), 0, 0, time.Local)
	}

	now := time.Now()

	return now.After(startDateTime) && now.Before(endDateTime), nil
}
