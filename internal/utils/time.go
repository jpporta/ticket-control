package utils

import "time"

func isWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

func IsLastWeekdayMonth(fn func()) {
	t := time.Now()
	m := t.Month()
	if isWeekend(t) {
		return
	}
	for t.Month() == m {
		t = t.AddDate(0, 0, 1)
		if !isWeekend(t) {
			return
		}
	}
	fn()
}

func IsLastWorkdayToMiddle(fn func()) {
	if isLastWorkdayUpTo(15) {
		fn()
	}
}
func IsLastWorkdayTo10(fn func()) {
	if isLastWorkdayUpTo(10) {
		fn()
	}
}

func isLastWorkdayUpTo(day int) bool {
	t := time.Now()
	if isWeekend(t) {
		return false
	}
	if t.Day() >= day {
		return true
	}
	for i := t.Day(); i < day; i++ {
		t = t.AddDate(0, 0, 1)
		if !isWeekend(t) {
			return false
		}
	}
	return true
}
