package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

func init() {
	// Vaud: Monday after the third Sunday of September.
	RegisterMethod("ch_vd_lundi_du_jeune_federal", func(a MethodArgs) (time.Time, error) {
		firstSun := firstSundayOf(a.Year, time.September)
		return firstSun.AddDate(0, 0, 15), nil
	})
	// Geneva: Thursday after the first Sunday of September.
	RegisterMethod("ch_ge_jeune_genevois", func(a MethodArgs) (time.Time, error) {
		return firstSundayOf(a.Year, time.September).AddDate(0, 0, 4), nil
	})
	// Glarus: First Thursday of April. If it falls in the week before Easter,
	// shift by one week.
	RegisterMethod("ch_gl_naefelser_fahrt", func(a MethodArgs) (time.Time, error) {
		thu := calc.DayOfMonth(a.Year, time.April, 1, time.Thursday)
		easter := calc.Easter(a.Year)
		if thu.Equal(easter.AddDate(0, 0, -3)) {
			thu = thu.AddDate(0, 0, 7)
		}
		return thu, nil
	})
	// Bern: fourth Monday of November.
	RegisterMethod("ch_be_zibelemaerit", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.November, 4, time.Monday), nil
	})
}

func firstSundayOf(year int, month time.Month) time.Time {
	return calc.DayOfMonth(year, month, 1, time.Sunday)
}
