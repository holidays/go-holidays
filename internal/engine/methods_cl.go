package engine

import "time"

func init() {
	// Chile: roll June 29 (St Peter & St Paul) to the nearest Monday.
	// Tue/Wed/Thu roll back; Fri rolls forward; weekend/Mon stay put.
	RegisterMethod("st_peter_st_paul_cl", func(a MethodArgs) (time.Time, error) {
		return rollClNearestMonday(a.Year, time.June, 29), nil
	})
	// Chile: roll October 12 (Columbus Day) to the nearest Monday, same rule.
	RegisterMethod("columbus_day_cl", func(a MethodArgs) (time.Time, error) {
		return rollClNearestMonday(a.Year, time.October, 12), nil
	})
	// Chile: Day of the Other Churches — Oct 31. Tue rolls back to Fri,
	// Wed rolls forward to Fri; other days unchanged.
	RegisterMethod("other_churches_day_cl", func(a MethodArgs) (time.Time, error) {
		d := time.Date(a.Year, time.October, 31, 0, 0, 0, 0, time.UTC)
		switch d.Weekday() {
		case time.Tuesday:
			return d.AddDate(0, 0, -4), nil
		case time.Wednesday:
			return d.AddDate(0, 0, 2), nil
		}
		return d, nil
	})
	// Chile: Día Nacional de los Pueblos Indígenas falls on the winter
	// solstice, except in 2021 when the law fixed it to June 21.
	RegisterMethod("indigenous_peoples_day_cl", func(a MethodArgs) (time.Time, error) {
		if a.Year == 2021 {
			return time.Date(2021, time.June, 21, 0, 0, 0, 0, time.UTC), nil
		}
		return solarTermDay(a.Year, time.June, clWinterSolsticeBase)
	})
}

// clWinterSolsticeBase is the June solstice constant, tuned to Chile time.
var clWinterSolsticeBase = solarTermBase{ranges: []solarTermRange{
	{2022, 2099, 21.019},
}}

func rollClNearestMonday(year int, month time.Month, day int) time.Time {
	d := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	switch d.Weekday() {
	case time.Tuesday, time.Wednesday, time.Thursday:
		return d.AddDate(0, 0, -(int(d.Weekday()) - 1))
	case time.Friday:
		return d.AddDate(0, 0, 3)
	}
	return d
}
