package pkg

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func today(loc *time.Location) time.Time {
	now := time.Now().In(loc)
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
}

func TestParseDate_keywords(t *testing.T) {
	t.Parallel()
	loc := DefaultLocation
	base := today(loc)

	tests := []struct {
		in   string
		want time.Time
	}{
		{"", base},
		{"today", base},
		{"Today", base},
		{"yesterday", base.AddDate(0, 0, -1)},
		{"tomorrow", base.AddDate(0, 0, 1)},
	}
	for _, tt := range tests {
		got, err := ParseDate(tt.in, loc)
		assert.NoError(t, err, tt.in)
		assert.Equal(t, tt.want, got, tt.in)
	}
}

func TestParseDate_weekdays(t *testing.T) {
	t.Parallel()
	loc := DefaultLocation
	base := today(loc)

	for _, in := range []string{"friday", "Friday", "fri", "monday", "mon", "sunday", "sat"} {
		got, err := ParseDate(in, loc)
		assert.NoError(t, err, in)
		diff := int(base.Sub(got).Hours() / 24)
		assert.GreaterOrEqual(t, diff, 0, in)
		assert.LessOrEqual(t, diff, 6, in)
	}

	got, err := ParseDate("friday", loc)
	assert.NoError(t, err)
	assert.Equal(t, time.Friday, got.Weekday())

	// A weekday name matching today resolves to today, not last week.
	sameDay := base.Weekday().String()
	got, err = ParseDate(sameDay, loc)
	assert.NoError(t, err)
	assert.Equal(t, base, got)
}

func TestParseDate_ordinals(t *testing.T) {
	t.Parallel()
	loc := DefaultLocation
	base := today(loc)

	got, err := ParseDate("1st", loc)
	assert.NoError(t, err)
	assert.Equal(t, 1, got.Day())
	assert.False(t, got.After(base))

	// A day later in the month than today resolves to a previous month.
	for _, in := range []string{"2nd", "15th", "23rd", "31st"} {
		got, err := ParseDate(in, loc)
		assert.NoError(t, err, in)
		assert.False(t, got.After(base), in)
		assert.LessOrEqual(t, base.Sub(got).Hours(), float64(24*93), in)
	}

	_, err = ParseDate("32nd", loc)
	assert.Error(t, err)
	_, err = ParseDate("0th", loc)
	assert.Error(t, err)
}

func TestParseDate_monthDay(t *testing.T) {
	t.Parallel()
	loc := DefaultLocation
	base := today(loc)

	for _, in := range []string{"jan 5", "Jan 5", "5 jan", "january 5", "5 january", "jan 5th"} {
		got, err := ParseDate(in, loc)
		assert.NoError(t, err, in)
		assert.Equal(t, time.January, got.Month(), in)
		assert.Equal(t, 5, got.Day(), in)
		assert.False(t, got.After(base), in)
	}
}

func TestParseDate_explicitFormats(t *testing.T) {
	t.Parallel()
	loc := DefaultLocation
	want := time.Date(2026, time.January, 5, 0, 0, 0, 0, loc)

	for _, in := range []string{"2026-01-05", "05-01-2026", "Jan 5, 2026", "jan 5, 2026", "January 5, 2026", "5 jan 2026", "jan 5th 2026"} {
		got, err := ParseDate(in, loc)
		assert.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}

	_, err := ParseDate("not a date", loc)
	assert.Error(t, err)
}

func TestParseTime_formats(t *testing.T) {
	t.Parallel()
	loc := DefaultLocation

	tests := []struct {
		in       string
		wantHour int
		wantMin  int
	}{
		{"5pm", 17, 0},
		{"5 pm", 17, 0},
		{"5PM", 17, 0},
		{"11 PM", 23, 0},
		{"5", 5, 0},
		{"17", 17, 0},
		{"5:30pm", 17, 30},
		{"5:30PM", 17, 30},
		{"15:04", 15, 4},
		{"3:04 pm", 15, 4},
		{"noon", 12, 0},
		{"evening", 18, 0},
		{"midnight", 0, 0},
	}
	for _, tt := range tests {
		got, err := ParseTime(tt.in, loc)
		assert.NoError(t, err, tt.in)
		assert.Equal(t, tt.wantHour, got.Hour(), tt.in)
		assert.Equal(t, tt.wantMin, got.Minute(), tt.in)
	}

	_, err := ParseTime("25:00", loc)
	assert.Error(t, err)
}
