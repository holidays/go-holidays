package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

func init() {
	RegisterMethod("christmas_eve_holiday", func(a MethodArgs) (time.Time, error) {
		switch a.Date.Weekday() {
		case time.Saturday:
			return a.Date.AddDate(0, 0, -1), nil
		case time.Sunday:
			return a.Date.AddDate(0, 0, -2), nil
		}
		return a.Date, nil
	})

	RegisterMethod("rosh_hashanah", func(a MethodArgs) (time.Time, error) {
		dates := map[int]time.Time{
			2014: time.Date(2014, 9, 25, 0, 0, 0, 0, time.UTC),
			2015: time.Date(2015, 9, 14, 0, 0, 0, 0, time.UTC),
			2016: time.Date(2016, 10, 3, 0, 0, 0, 0, time.UTC),
			2017: time.Date(2017, 9, 21, 0, 0, 0, 0, time.UTC),
			2018: time.Date(2018, 9, 10, 0, 0, 0, 0, time.UTC),
			2019: time.Date(2019, 9, 30, 0, 0, 0, 0, time.UTC),
			2020: time.Date(2020, 9, 19, 0, 0, 0, 0, time.UTC),
		}
		return dates[a.Year], nil
	})

	RegisterMethod("yom_kippur", func(a MethodArgs) (time.Time, error) {
		dates := map[int]time.Time{
			2014: time.Date(2014, 10, 4, 0, 0, 0, 0, time.UTC),
			2015: time.Date(2015, 9, 23, 0, 0, 0, 0, time.UTC),
			2016: time.Date(2016, 10, 12, 0, 0, 0, 0, time.UTC),
			2017: time.Date(2017, 9, 30, 0, 0, 0, 0, time.UTC),
			2018: time.Date(2018, 9, 19, 0, 0, 0, 0, time.UTC),
			2019: time.Date(2019, 10, 9, 0, 0, 0, 0, time.UTC),
			2020: time.Date(2020, 9, 28, 0, 0, 0, 0, time.UTC),
		}
		return dates[a.Year], nil
	})

	// Georgia State Holiday: Monday on or before April 26.
	RegisterMethod("georgia_state_holiday", func(a MethodArgs) (time.Time, error) {
		d := time.Date(a.Year, time.Month(a.Month), 26, 0, 0, 0, 0, time.UTC)
		for d.Weekday() != time.Monday {
			d = d.AddDate(0, 0, -1)
		}
		return d, nil
	})

	// Lee-Jackson Day: Friday before the third Monday in January.
	RegisterMethod("lee_jackson_day", func(a MethodArgs) (time.Time, error) {
		thirdMonday := calc.DayOfMonth(a.Year, time.Month(a.Month), 3, time.Monday)
		d := thirdMonday
		for d.Weekday() != time.Friday {
			d = d.AddDate(0, 0, -1)
		}
		return d, nil
	})

	// Election Day: Tuesday after the first Monday of November.
	RegisterMethod("election_day", func(a MethodArgs) (time.Time, error) {
		firstMonday := calc.DayOfMonth(a.Year, time.November, 1, time.Monday)
		return firstMonday.AddDate(0, 0, 1), nil
	})

	// Same as election_day but only in even years.
	RegisterMethod("even_year_election_day", func(a MethodArgs) (time.Time, error) {
		if a.Year%2 != 0 {
			return time.Time{}, nil
		}
		firstMonday := calc.DayOfMonth(a.Year, time.November, 1, time.Monday)
		return firstMonday.AddDate(0, 0, 1), nil
	})

	// Inauguration Day: January 20 in years following a presidential election
	// (i.e. year % 4 == 1: 2017, 2021, 2025, ...).
	RegisterMethod("us_inauguration_day", func(a MethodArgs) (time.Time, error) {
		if a.Year%4 != 1 {
			return time.Time{}, nil
		}
		return time.Date(a.Year, time.January, 20, 0, 0, 0, 0, time.UTC), nil
	})

	// Day after Thanksgiving: day after the 4th Thursday of November.
	RegisterMethod("day_after_thanksgiving", func(a MethodArgs) (time.Time, error) {
		thanksgiving := calc.DayOfMonth(a.Year, time.November, 4, time.Thursday)
		return thanksgiving.AddDate(0, 0, 1), nil
	})

	// Juneteenth's observed date is region-dependent, so this reads a.Region
	// (the region the caller requested; see requestedRegion in resolve.go).
	// Utah moves it to a Monday from either direction: Tue-Fri back to the
	// preceding Monday, Sat/Sun forward to the next one. Everywhere else it
	// behaves as to_weekday_if_weekend.
	RegisterMethod("juneteenth_national_independence_day", func(a MethodArgs) (time.Time, error) {
		w := a.Date.Weekday()
		if a.Region == "us_ut" {
			switch w {
			case time.Sunday:
				return a.Date.AddDate(0, 0, 1), nil
			case time.Saturday:
				return a.Date.AddDate(0, 0, 2), nil
			case time.Monday:
				return a.Date, nil
			default:
				return a.Date.AddDate(0, 0, -(int(w) - 1)), nil
			}
		}
		switch w {
		case time.Sunday:
			return a.Date.AddDate(0, 0, 1), nil
		case time.Saturday:
			return a.Date.AddDate(0, 0, -1), nil
		default:
			return a.Date, nil
		}
	})
}
