package utils

import (
	"time"
)

func isWeekend(t time.Time) bool {
	return t.Weekday() == time.Saturday || t.Weekday() == time.Sunday
}

func IsDayLastWorkDayMonth(t time.Time) bool {
	m := t.Month()
	if isWeekend(t) {
		return false
	}
	t = t.AddDate(0, 0, 1)
	for t.Month() == m {
		if !isWeekend(t) {
			return false
		}
		t = t.AddDate(0, 0, 1)
	}
	return true
}

func IsLastWeekdayMonth(fn func()) {
	if IsDayLastWorkDayMonth(time.Now()) {
		fn()
	}
}

func IsLastWorkdayToMiddle(fn func()) {
	if isLastWorkdayUpTo(time.Now(), 15) {
		fn()
	}
}
func IsLastWorkdayTo10(fn func()) {
	if isLastWorkdayUpTo(time.Now(), 10) {
		fn()
	}
}

func isLastWorkdayUpTo(t time.Time, day int) bool {
	if isWeekend(t) {
		return false
	}
	if t.Day() == day {
		return true
	}
	t = t.AddDate(0, 0, 1)
	for t.Day() <= day {
		if !isWeekend(t) {
			return false
		}
		t = t.AddDate(0, 0, 1)
	}
	return true
}
