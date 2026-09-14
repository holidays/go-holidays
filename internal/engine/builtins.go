package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

// The 15 well-known methods from holidays/definitions/METHODS.yml.
// calculate_day_of_month is intentionally omitted from the registry: it is only
// ever an internal helper, never a YAML function: reference. If a future
// definition references it, register it then.
func init() {
	RegisterMethod("easter", func(a MethodArgs) (time.Time, error) {
		return calc.Easter(a.Year), nil
	})
	RegisterMethod("orthodox_easter", func(a MethodArgs) (time.Time, error) {
		return calc.OrthodoxEaster(a.Year), nil
	})
	RegisterMethod("orthodox_easter_julian", func(a MethodArgs) (time.Time, error) {
		return calc.OrthodoxEasterJulian(a.Year), nil
	})
	RegisterMethod("to_monday_if_sunday", func(a MethodArgs) (time.Time, error) {
		return calc.ToMondayIfSunday(a.Date), nil
	})
	RegisterMethod("to_monday_if_weekend", func(a MethodArgs) (time.Time, error) {
		return calc.ToMondayIfWeekend(a.Date), nil
	})
	RegisterMethod("to_weekday_if_weekend", func(a MethodArgs) (time.Time, error) {
		return calc.ToWeekdayIfWeekend(a.Date), nil
	})
	RegisterMethod("to_weekday_if_boxing_weekend", func(a MethodArgs) (time.Time, error) {
		return calc.ToWeekdayIfBoxingWeekend(a.Date), nil
	})
	RegisterMethod("to_weekday_if_boxing_weekend_from_year", func(a MethodArgs) (time.Time, error) {
		// Defined as to_tuesday_if_sunday_or_monday_if_saturday(Dec26):
		// Sat->+2, Sun->+2, Mon unchanged. (au_tas/au_nt Boxing Day)
		boxing := time.Date(a.Year, time.December, 26, 0, 0, 0, 0, time.UTC)
		return calc.ToTuesdayIfSundayOrMondayIfSaturday(boxing), nil
	})
	RegisterMethod("to_weekday_if_boxing_weekend_from_year_or_to_tuesday_if_monday", func(a MethodArgs) (time.Time, error) {
		// Despite the name, this is defined as to_weekday_if_boxing_weekend(Dec26):
		// Sat->+2, Sun->+2, Mon->+1. (au_sa Proclamation Day)
		boxing := time.Date(a.Year, time.December, 26, 0, 0, 0, 0, time.UTC)
		return calc.ToWeekdayIfBoxingWeekend(boxing), nil
	})
	RegisterMethod("to_tuesday_if_sunday_or_monday_if_saturday", func(a MethodArgs) (time.Time, error) {
		return calc.ToTuesdayIfSundayOrMondayIfSaturday(a.Date), nil
	})
	RegisterMethod("to_the_weekday_after", func(a MethodArgs) (time.Time, error) {
		return calc.ToTheWeekdayAfter(a.Date), nil
	})
	RegisterMethod("to_the_second_weekday_after", func(a MethodArgs) (time.Time, error) {
		return calc.ToTheSecondWeekdayAfter(a.Date), nil
	})
	RegisterMethod("to_previous_day_if_leap_year", func(a MethodArgs) (time.Time, error) {
		return calc.ToPreviousDayIfLeapYear(a.Date), nil
	})
	RegisterMethod("lunar_to_solar", func(a MethodArgs) (time.Time, error) {
		return calc.LunarToSolar(a.Year, a.Month, a.Day, a.Region)
	})
}
