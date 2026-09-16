package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

// Hijri month numbers used to locate Ramazan Bayramı (Eid al-Fitr) and
// Kurban Bayramı (Eid al-Adha).
const (
	trShawwal     = 10
	trDhuAlHijjah = 12
)

// trFeastKey identifies a Diyanet override entry.
type trFeastKey struct {
	method string
	year   int
}

// trDiyanetOverrides holds proclaimed first days that differ from the pure
// arithmetic Hijri calculation, keyed by (method, Gregorian year). Ported
// verbatim from Holidays::Definition::CustomMethods::TR::DIYANET_OVERRIDES
// in the definitions repo (definitions PR #385, "tr-hijri-date-calculator").
// Years with no entry fall back to calc.HijriYearOccurrence.
var trDiyanetOverrides = map[trFeastKey]time.Time{
	{"ramadan_feast", 2014}: time.Date(2014, time.July, 28, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2015}: time.Date(2015, time.July, 17, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2016}: time.Date(2016, time.July, 5, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2017}: time.Date(2017, time.June, 25, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2019}: time.Date(2019, time.June, 4, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2022}: time.Date(2022, time.May, 2, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2023}: time.Date(2023, time.April, 21, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2025}: time.Date(2025, time.March, 30, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2027}: time.Date(2027, time.March, 9, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2028}: time.Date(2028, time.February, 26, 0, 0, 0, 0, time.UTC),
	{"ramadan_feast", 2030}: time.Date(2030, time.February, 4, 0, 0, 0, 0, time.UTC),

	{"sacrifice_feast", 2014}: time.Date(2014, time.October, 4, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2016}: time.Date(2016, time.September, 12, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2017}: time.Date(2017, time.September, 1, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2018}: time.Date(2018, time.August, 21, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2019}: time.Date(2019, time.August, 11, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2022}: time.Date(2022, time.July, 9, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2023}: time.Date(2023, time.June, 28, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2024}: time.Date(2024, time.June, 16, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2025}: time.Date(2025, time.June, 6, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2027}: time.Date(2027, time.May, 16, 0, 0, 0, 0, time.UTC),
	{"sacrifice_feast", 2030}: time.Date(2030, time.April, 13, 0, 0, 0, 0, time.UTC),
}

// Ramazan Bayramı (Ramadan/Sacrifice feasts) are tied to the Hijri calendar.
// Both are derived from the arithmetic Islamic calendar (calc package). The
// civil holidays are proclaimed by Türkiye's Diyanet and in most years so
// far have landed a day (2016 Ramazan Bayramı, two days) before the
// arithmetic result; trDiyanetOverrides pins those years to the proclaimed
// date. Years with no entry fall back to the calculation. See
// definitions#377 / PR #385.
func init() {
	RegisterMethod("ramadan_feast", func(a MethodArgs) (time.Time, error) {
		return trFeastDate("ramadan_feast", a.Year, trShawwal, 1)
	})
	RegisterMethod("sacrifice_feast", func(a MethodArgs) (time.Time, error) {
		return trFeastDate("sacrifice_feast", a.Year, trDhuAlHijjah, 10)
	})
}

// trFeastDate returns the Gregorian date of the given Hijri month/day
// occurrence in year, preferring a Diyanet-proclaimed override when one is
// recorded. If neither an override nor a calculated occurrence exists for
// year (the arithmetic calendar's fixed +/-1-year search window can miss an
// occurrence near its edges, e.g. 1970-1973 for Dhu al-Hijjah 10 - Ruby's
// HijriDate#gregorian_year_occurrence has the same gap), it returns the zero
// Time with a nil error: the same "no holiday this year" convention used by
// even_year_election_day in methods_us.go, not a hard failure. Ruby's
// CustomMethods::TR#feast likewise returns nil in this case and the gem
// treats that as no holiday, not an error.
func trFeastDate(method string, year, hijriMonth, hijriDay int) (time.Time, error) {
	if d, ok := trDiyanetOverrides[trFeastKey{method, year}]; ok {
		return d, nil
	}
	d, ok := calc.HijriYearOccurrence(year, hijriMonth, hijriDay)
	if !ok {
		return time.Time{}, nil
	}
	return d, nil
}
