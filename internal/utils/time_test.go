package utils

import (
	"testing"
	"time"
)

func Test_IsDayLastWorkDayMonth(t *testing.T) {
	// Thursday, August 28, 2025
	res := IsDayLastWorkDayMonth(
		time.Date(2025, 8, 28, 0, 0, 0, 0, time.UTC),
	)
	if res {
		t.Fatal("Expected false, got true")
	}
	// Friday, August 29, 2025
	res = IsDayLastWorkDayMonth(
		time.Date(2025, 8, 29, 0, 0, 0, 0, time.UTC),
	)
	if !res {
		t.Fatal("Expected true, got false")
	}
	// Saturday, August 30, 2025
	res = IsDayLastWorkDayMonth(
		time.Date(2025, 8, 30, 0, 0, 0, 0, time.UTC),
	)
	if res {
		t.Fatal("Expected false, got true")
	}
	// Sunday, August 31, 2025
	res = IsDayLastWorkDayMonth(
		time.Date(2025, 8, 31, 0, 0, 0, 0, time.UTC),
	)
	if res {
		t.Fatal("Expected false, got true")
	}
}

func Test_isLastWorkdayUpTo(t *testing.T) {
	res := isLastWorkdayUpTo(
		time.Date(2025, 8, 7, 0, 0, 0, 0, time.UTC), 10)
	if res {
		t.Fatal("Aug 7 is not the last workday up to 10")
	}
	res = isLastWorkdayUpTo(
		time.Date(2025, 8, 8, 0, 0, 0, 0, time.UTC), 10)
	if !res {
		t.Fatal("Aug 8 is the last workday up to 10")
	}
	res = isLastWorkdayUpTo(
		time.Date(2025, 8, 14, 0, 0, 0, 0, time.UTC), 15)
	if res {
		t.Fatal("Aug 14 is not the last workday up to 15")
	}
	res = isLastWorkdayUpTo(
		time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC), 15)
	if !res {
		t.Fatal("Aug 15 is the last workday up to 15")
	}
	res = isLastWorkdayUpTo(
		time.Date(2025, 8, 22, 0, 0, 0, 0, time.UTC), 25)
	if res {
		t.Fatal("Aug 22 is not the last workday up to 25")
	}
	res = isLastWorkdayUpTo(
		time.Date(2025, 8, 24, 0, 0, 0, 0, time.UTC), 25)
	if res {
		t.Fatal("Aug 24 is not the last workday up to 25")
	}
	res = isLastWorkdayUpTo(
		time.Date(2025, 8, 25, 0, 0, 0, 0, time.UTC), 25)
	if !res {
		t.Fatal("Aug 25 is the last workday up to 25")
	}
}
