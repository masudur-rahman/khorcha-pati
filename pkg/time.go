package pkg

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
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

// LoadTimezone returns the location for tz, falling back to the default one.
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

// StartOfMonth returns midnight on the first day of the current month.
func StartOfMonth(loc *time.Location) time.Time {
	if loc == nil {
		loc = DefaultLocation
	}

	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
}

var (
	// ordinalDayRe matches a bare day-of-month ordinal like "1st" or "23rd".
	ordinalDayRe = regexp.MustCompile(`^(\d{1,2})(?:st|nd|rd|th)$`)
	// ordinalSuffixRe strips ordinal suffixes inside longer dates ("jan 5th").
	ordinalSuffixRe = regexp.MustCompile(`(?i)\b(\d{1,2})(?:st|nd|rd|th)\b`)

	dateFormats = []string{
		time.DateOnly, "02-01-2006",
		"Jan 2, 2006", "January 2, 2006", "Jan 2 2006", "January 2 2006",
		"2 Jan 2006", "2 January 2006",
	}
	// monthDayFormats are year-less dates resolved to the most recent occurrence.
	monthDayFormats = []string{"Jan 2", "January 2", "2 Jan", "2 January"}

	weekdayNames = map[string]time.Weekday{
		"sunday": time.Sunday, "sun": time.Sunday,
		"monday": time.Monday, "mon": time.Monday,
		"tuesday": time.Tuesday, "tue": time.Tuesday, "tues": time.Tuesday,
		"wednesday": time.Wednesday, "wed": time.Wednesday,
		"thursday": time.Thursday, "thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday,
		"friday": time.Friday, "fri": time.Friday,
		"saturday": time.Saturday, "sat": time.Saturday,
	}

	timeFormats = []string{
		time.TimeOnly, time.Kitchen, "3:04pm", "3:04 PM", "3:04 pm", "3:04", "15:04",
		"3PM", "3pm", "3 PM", "3 pm", "15", "3",
	}

	// namedHours maps spoken times of day to a representative hour.
	namedHours = map[string]int{
		"midnight": 0, "morning": 6, "noon": 12, "afternoon": 15, "evening": 18, "night": 22,
	}
)

// IsNamedHour reports whether w (lowercase) is a named time of day.
func IsNamedHour(w string) bool {
	_, ok := namedHours[w]
	return ok
}

// IsFullWeekdayName reports whether w (lowercase) is a full weekday name.
func IsFullWeekdayName(w string) bool {
	switch w {
	case "sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday":
		return true
	}
	return false
}

// ParseDate parses keywords (today/yesterday/tomorrow), weekday names,
// day-of-month ordinals ("1st"), year-less month-days ("jan 5") and explicit
// date formats. Relative forms resolve to the most recent past occurrence.
func ParseDate(date string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = DefaultLocation
	}
	today := startOfToday(loc)

	lower := strings.ToLower(date)
	switch lower {
	case "", "today":
		return today, nil
	case "yesterday":
		return today.AddDate(0, 0, -1), nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	}
	if wd, ok := weekdayNames[lower]; ok {
		return lastWeekday(today, wd), nil
	}
	if m := ordinalDayRe.FindStringSubmatch(lower); m != nil {
		day, _ := strconv.Atoi(m[1])
		return mostRecentMonthDay(today, day)
	}

	return parseDateFormats(normalizeDate(date), today, loc)
}

// startOfToday returns midnight today in loc.
func startOfToday(loc *time.Location) time.Time {
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

// normalizeDate capitalizes words (Go's layout parsing is case-sensitive for
// month names) and strips ordinal suffixes so "jan 5th" parses as "Jan 5".
func normalizeDate(s string) string {
	s = ordinalSuffixRe.ReplaceAllString(s, "$1")
	words := strings.Fields(s)
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
	return strings.Join(words, " ")
}

// lastWeekday returns the most recent occurrence of wd, counting today.
func lastWeekday(today time.Time, wd time.Weekday) time.Time {
	diff := (int(today.Weekday()) - int(wd) + 7) % 7
	return today.AddDate(0, 0, -diff)
}

// mostRecentMonthDay returns the latest date (today or earlier) whose
// day-of-month equals day, skipping months that are too short.
func mostRecentMonthDay(today time.Time, day int) (time.Time, error) {
	const maxDayOfMonth = 31
	if day < 1 || day > maxDayOfMonth {
		return time.Time{}, fmt.Errorf("invalid day of month: %d", day)
	}
	for back := 0; back < 3; back++ {
		t := time.Date(today.Year(), today.Month()-time.Month(back), day, 0, 0, 0, 0, today.Location())
		if t.Day() == day && !t.After(today) {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid day of month: %d", day)
}

// parseDateFormats tries explicit formats, then year-less month-days resolved
// to the most recent past occurrence.
func parseDateFormats(date string, today time.Time, loc *time.Location) (time.Time, error) {
	for _, format := range dateFormats {
		if t, err := time.ParseInLocation(format, date, loc); err == nil {
			return t, nil
		}
	}
	for _, format := range monthDayFormats {
		t, err := time.ParseInLocation(format, date, loc)
		if err != nil {
			continue
		}
		t = time.Date(today.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
		if t.After(today) {
			t = t.AddDate(-1, 0, 0)
		}
		return t, nil
	}
	return time.Time{}, fmt.Errorf("invalid date format")
}

// ParseTime parses named times of day (morning, noon, night, ...) and clock
// formats, including hour-only forms like "5pm" or "17".
func ParseTime(tim string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = DefaultLocation
	}

	now := time.Now().In(loc)
	lower := strings.ToLower(tim)
	if lower == "" || lower == "now" {
		return now, nil
	}
	if hour, ok := namedHours[lower]; ok {
		return time.Date(now.Year(), now.Month(), now.Day(), hour, 0, 0, 0, loc), nil
	}

	for _, format := range timeFormats {
		if t, err := time.ParseInLocation(format, tim, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid time format")
}
