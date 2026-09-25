// Package clock provides the app's notion of "now" in the configured time zone
// plus helpers for the YYYY-MM-DD / YYYY-MM string formats used everywhere.
package clock

import (
	"fmt"
	"time"
)

const (
	DateLayout  = "2006-01-02"
	MonthLayout = "2006-01"
)

// Clock returns the current time in the app's location. Inject a Fixed clock in tests.
type Clock interface {
	Now() time.Time
}

type systemClock struct{ loc *time.Location }

// System returns a Clock reporting wall time in loc.
func System(loc *time.Location) Clock { return systemClock{loc: loc} }

func (c systemClock) Now() time.Time { return time.Now().In(c.loc) }

// Fixed is a Clock frozen at T. For tests.
type Fixed struct{ T time.Time }

func (f Fixed) Now() time.Time { return f.T }

// Today returns the current local date as YYYY-MM-DD.
func Today(c Clock) string { return c.Now().Format(DateLayout) }

// CurrentMonth returns the current local month as YYYY-MM.
func CurrentMonth(c Clock) string { return c.Now().Format(MonthLayout) }

// ParseDate validates a YYYY-MM-DD string.
func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse(DateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q, want YYYY-MM-DD", s)
	}
	return t, nil
}

// ParseMonth validates a YYYY-MM string and returns the first day of that month (UTC).
func ParseMonth(s string) (time.Time, error) {
	t, err := time.Parse(MonthLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month %q, want YYYY-MM", s)
	}
	return t, nil
}

// MonthOf returns the YYYY-MM month of a YYYY-MM-DD date string (no validation).
func MonthOf(date string) string {
	if len(date) < 7 {
		return date
	}
	return date[:7]
}

// AddMonths shifts a YYYY-MM month by n (may be negative).
func AddMonths(month string, n int) (string, error) {
	t, err := ParseMonth(month)
	if err != nil {
		return "", err
	}
	return t.AddDate(0, n, 0).Format(MonthLayout), nil
}

// MonthRange returns the first and last dates (YYYY-MM-DD, inclusive) of a YYYY-MM month.
func MonthRange(month string) (first, last string, err error) {
	t, err := ParseMonth(month)
	if err != nil {
		return "", "", err
	}
	return t.Format(DateLayout), t.AddDate(0, 1, -1).Format(DateLayout), nil
}

// DaysIn returns the number of days in a YYYY-MM month.
func DaysIn(month string) (int, error) {
	t, err := ParseMonth(month)
	if err != nil {
		return 0, err
	}
	return t.AddDate(0, 1, -1).Day(), nil
}

// MonthsBetween lists months from..to inclusive (YYYY-MM). Empty if from > to.
func MonthsBetween(from, to string) ([]string, error) {
	f, err := ParseMonth(from)
	if err != nil {
		return nil, err
	}
	t, err := ParseMonth(to)
	if err != nil {
		return nil, err
	}
	var out []string
	for m := f; !m.After(t); m = m.AddDate(0, 1, 0) {
		out = append(out, m.Format(MonthLayout))
	}
	return out, nil
}
