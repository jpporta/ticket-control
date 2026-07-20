package clock

import "time"

// Now returns the current wall-clock time in UTC.
// All domain code MUST read time through this function so quota windows and
// "today" checks are stable regardless of the server's local TZ.
func Now() time.Time {
	return time.Now().UTC()
}

// Today returns the UTC date for the current moment.
func Today() time.Time {
	t := Now()
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
