package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

func init() {
	RegisterMethod("afl_grand_final", func(a MethodArgs) (time.Time, error) {
		// AFL Grand Final is officially declared each year by the Victorian
		// government; known years are hard-coded, with a fallback to the last
		// Friday of September.
		switch a.Year {
		case 2015:
			return time.Date(2015, 10, 2, 0, 0, 0, 0, time.UTC), nil
		case 2016:
			return time.Date(2016, 9, 30, 0, 0, 0, 0, time.UTC), nil
		case 2017:
			return time.Date(2017, 9, 29, 0, 0, 0, 0, time.UTC), nil
		case 2020:
			return time.Date(2020, 10, 23, 0, 0, 0, 0, time.UTC), nil
		case 2022:
			return time.Date(2022, 9, 23, 0, 0, 0, 0, time.UTC), nil
		}
		return calc.DayOfMonth(a.Year, time.September, -1, time.Friday), nil
	})

	RegisterMethod("qld_queens_bday_october", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.October, 1, time.Monday), nil
	})

	RegisterMethod("qld_kings_bday_october", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.October, 1, time.Monday), nil
	})

	RegisterMethod("qld_queens_birthday_june", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.June, 2, time.Monday), nil
	})

	RegisterMethod("qld_labour_day_may", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.May, 1, time.Monday), nil
	})

	RegisterMethod("qld_labour_day_october", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.October, 1, time.Monday), nil
	})

	// Hobart Show Day: Thursday before the 4th Saturday of October.
	RegisterMethod("hobart_show_day", func(a MethodArgs) (time.Time, error) {
		fourthSat := calc.DayOfMonth(a.Year, time.October, 4, time.Saturday)
		return fourthSat.AddDate(0, 0, -2), nil
	})

	RegisterMethod("march_pub_hol_sa", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.March, 2, time.Monday), nil
	})

	RegisterMethod("may_pub_hol_sa", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.May, 3, time.Monday), nil
	})

	// Brisbane Ekka holiday: Wednesday during the RNA Show. The RNA show
	// starts the first Friday of August unless that's before August 5, in
	// which case it starts the second Friday. Show Day is the following
	// Wednesday (+5 days from the Friday start).
	RegisterMethod("qld_brisbane_ekka_holiday", func(a MethodArgs) (time.Time, error) {
		first := calc.DayOfMonth(a.Year, time.August, 1, time.Friday)
		if first.Day() < 5 {
			second := calc.DayOfMonth(a.Year, time.August, 2, time.Friday)
			return second.AddDate(0, 0, 5), nil
		}
		return first.AddDate(0, 0, 5), nil
	})

	// to_nearest_monday_after: roll forward to the next Monday-on-or-after.
	RegisterMethod("to_nearest_monday_after", func(a MethodArgs) (time.Time, error) {
		switch w := int(a.Date.Weekday()); w {
		case 0:
			return a.Date.AddDate(0, 0, 1), nil
		case 1:
			return a.Date, nil
		default:
			return a.Date.AddDate(0, 0, 8-w), nil
		}
	})
}
