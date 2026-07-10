package pkg

import (
	"fmt"
	"time"
)

var DefaultLocation *time.Location

func init() {
	loc, err := time.LoadLocation("Asia/Dhaka")
	if err != nil {
		DefaultLocation = time.UTC
	} else {
		DefaultLocation = loc
	}
}

func LoadTimezone(tz string) *time.Location {
	if tz == "" {
		tz = "Asia/Dhaka"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return DefaultLocation
	}
	return loc
}

func StartOfMonth(loc *time.Location) time.Time {
	if loc == nil {
		loc = DefaultLocation
	}

	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
}

func ParseDate(date string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = DefaultLocation
	}

	now := time.Now().In(loc)
	now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	switch date {
	case "", "today", "Today":
		return now, nil
	case "yesterday", "Yesterday":
		return now.AddDate(0, 0, -1), nil
	case "tomorrow", "Tomorrow":
		return now.AddDate(0, 0, 1), nil
	}

	dateFormats := []string{time.DateOnly, "02-01-2006", "Jan 2, 2006", "January 2, 2006"}
	for _, format := range dateFormats {
		t, err := time.ParseInLocation(format, date, loc)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format")
}

func ParseTime(tim string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = DefaultLocation
	}

	now := time.Now().In(loc)
	switch tim {
	case "", "now", "Now":
		return now, nil
	case "midnight", "Midnight":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc), nil
	case "morning", "Morning":
		return time.Date(now.Year(), now.Month(), now.Day(), 6, 0, 0, 0, loc), nil
	case "noon", "Noon":
		return time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc), nil
	case "afternoon", "Afternoon":
		return time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, loc), nil
	case "evening", "Evening":
		return time.Date(now.Year(), now.Month(), now.Day(), 18, 0, 0, 0, loc), nil
	case "night", "Night":
		return time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, loc), nil
	}

	timeFormats := []string{time.TimeOnly, time.Kitchen, "3:04pm", "3:04 PM", "3:04 pm", "3:04", "15:04"}
	for _, format := range timeFormats {
		t, err := time.ParseInLocation(format, tim, loc)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format")
}
