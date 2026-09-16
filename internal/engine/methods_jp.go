package engine

import (
	"fmt"
	"math"
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

func init() {
	// Vernal Equinox Day (春分の日): astronomical formula. Falls in March.
	RegisterMethod("jp_vernal_equinox_day", func(a MethodArgs) (time.Time, error) {
		return jpEquinox(a.Year, time.March, vernalEquinoxBase)
	})
	// Autumn Equinox Day (秋分の日): formula identical shape, different base
	// constants and falls in September. This method is named
	// `jp_national_culture_day` for historical reasons.
	RegisterMethod("jp_national_culture_day", func(a MethodArgs) (time.Time, error) {
		return jpEquinox(a.Year, time.September, autumnEquinoxBase)
	})

	RegisterMethod("jp_mountain_holiday", func(a MethodArgs) (time.Time, error) {
		return time.Date(a.Year, time.August, 11, 0, 0, 0, 0, time.UTC), nil
	})

	// Citizens' Holiday: the day before Autumn Equinox if that equinox is a Wednesday.
	RegisterMethod("jp_citizens_holiday", func(a MethodArgs) (time.Time, error) {
		eq, err := jpEquinox(a.Year, time.September, autumnEquinoxBase)
		if err != nil {
			return time.Time{}, err
		}
		if eq.Weekday() == time.Wednesday {
			return eq.AddDate(0, 0, -1), nil
		}
		return time.Time{}, nil
	})

	// Generic substitute: if the supplied date is a Sunday, return the next
	// non-Sunday non-holiday day; otherwise return zero (no substitute needed).
	RegisterMethod("jp_substitute_holiday", func(a MethodArgs) (time.Time, error) {
		if a.Date.IsZero() || a.Date.Weekday() != time.Sunday {
			return time.Time{}, nil
		}
		return jpNextWeekday(a.Date.AddDate(0, 0, 1), a.Year), nil
	})

	// Each named substitute wrapper computes the original holiday's date for
	// the year, then defers to the same Sunday-check + next-weekday logic.
	RegisterMethod("jp_marine_day_substitute", jpSubstituteWrapper(func(year int) time.Time {
		return calc.DayOfMonth(year, time.July, 3, time.Monday)
	}))
	RegisterMethod("jp_health_sports_day_substitute", jpSubstituteWrapper(func(year int) time.Time {
		return calc.DayOfMonth(year, time.October, 2, time.Monday)
	}))
	RegisterMethod("jp_respect_for_aged_holiday_substitute", jpSubstituteWrapper(func(year int) time.Time {
		return calc.DayOfMonth(year, time.September, 3, time.Monday)
	}))
	RegisterMethod("jp_mountain_holiday_substitute", jpSubstituteWrapper(func(year int) time.Time {
		return time.Date(year, time.August, 11, 0, 0, 0, 0, time.UTC)
	}))
	RegisterMethod("jp_vernal_equinox_day_substitute", jpSubstituteWrapper(func(year int) time.Time {
		d, _ := jpEquinox(year, time.March, vernalEquinoxBase)
		return d
	}))
	RegisterMethod("jp_national_culture_day_substitute", jpSubstituteWrapper(func(year int) time.Time {
		d, _ := jpEquinox(year, time.September, autumnEquinoxBase)
		return d
	}))
}

func jpSubstituteWrapper(originalDate func(year int) time.Time) Method {
	return func(a MethodArgs) (time.Time, error) {
		base := originalDate(a.Year)
		if base.IsZero() || base.Weekday() != time.Sunday {
			return time.Time{}, nil
		}
		return jpNextWeekday(base.AddDate(0, 0, 1), a.Year), nil
	}
}

// jpNextWeekday walks forward from date, skipping Sundays and any fixed-mday
// holiday registered under "jp" that applies in the given year, honoring each
// rule's year_ranges filter.
func jpNextWeekday(d time.Time, year int) time.Time {
	fixed := jpFixedHolidaysByMonth(year)
	for {
		if d.Weekday() == time.Sunday || fixed[int(d.Month())][d.Day()] {
			d = d.AddDate(0, 0, 1)
			continue
		}
		return d
	}
}

func jpFixedHolidaysByMonth(year int) map[int]map[int]bool {
	out := map[int]map[int]bool{}
	for _, r := range rulesForCountry("jp") {
		if r.Mday <= 0 || r.Function != "" || !r.AppliesIn(year) {
			continue
		}
		if out[r.Month] == nil {
			out[r.Month] = map[int]bool{}
		}
		out[r.Month][r.Mday] = true
	}
	return out
}

// Astronomical base constants, from jp.yaml.
type equinoxBase struct {
	ranges []struct {
		lo, hi int
		base   float64
	}
}

var vernalEquinoxBase = equinoxBase{ranges: []struct {
	lo, hi int
	base   float64
}{
	{1851, 1899, 19.8277},
	{1900, 1979, 20.8357},
	{1980, 2099, 20.8431},
	{2100, 2150, 21.8510},
}}

var autumnEquinoxBase = equinoxBase{ranges: []struct {
	lo, hi int
	base   float64
}{
	{1851, 1899, 22.2588},
	{1900, 1979, 23.2588},
	{1980, 2099, 23.2488},
	{2100, 2150, 24.2488},
}}

func jpEquinox(year int, month time.Month, eb equinoxBase) (time.Time, error) {
	var (
		base  float64
		found bool
	)
	for _, r := range eb.ranges {
		if year >= r.lo && year <= r.hi {
			base = r.base
			found = true
			break
		}
	}
	if !found {
		return time.Time{}, fmt.Errorf("jp equinox: year %d out of supported range", year)
	}
	diff := float64(year - 1980)
	day := math.Floor(base + 0.242194*diff - math.Floor(diff/4))
	return time.Date(year, month, int(day), 0, 0, 0, 0, time.UTC), nil
}
