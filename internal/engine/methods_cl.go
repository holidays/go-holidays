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

// clWinterSolsticeBase lets solarTermDay find the date of the June (winter)
// solstice in Chile.
//
// The 21.019 is the 1980 solstice expressed as a fractional day of June:
// very early on June 21. solarTermDay starts from that point, adds about a
// quarter day of drift per year, and subtracts a day for each leap year.
//
// The value is tied to Chilean local time, not UTC. The solstice can fall
// late on June 20 in Chile while it is already June 21 in UTC (in 2029, for
// example), and the holiday follows the Chilean date.
//
// The value comes from the upstream Ruby gem, where it was checked against
// feriados.cl and published solstice tables. It is only valid for 2022 through
// 2099; solarTermDay returns an error for any other year.
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
