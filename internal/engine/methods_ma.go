package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

// Hijri month numbers used to locate Eid al-Fitr and Eid al-Adha.
const (
	maShawwal     = 10
	maDhuAlHijjah = 12
)

// maFeastKey identifies a Ministry override entry.
type maFeastKey struct {
	method string
	year   int
}

// maMinistryOverrides holds proclaimed first days that differ from the pure
// arithmetic Hijri calculation, keyed by (method, Gregorian year). Ported
// verbatim from Holidays::Definition::CustomMethods::MA::MINISTRY_OVERRIDES
// in the definitions repo (definitions PR #391, "ma-hijri-date-calculator").
// Years with no entry fall back to calc.HijriYearOccurrence.
var maMinistryOverrides = map[maFeastKey]time.Time{
	{"eid_al_fitr", 2022}: time.Date(2022, time.May, 2, 0, 0, 0, 0, time.UTC),

	{"eid_al_adha", 2021}: time.Date(2021, time.July, 21, 0, 0, 0, 0, time.UTC),
}

// Eid al-Fitr (1 Shawwal) and Eid al-Adha (10 Dhul-Hijjah) are tied to the
// Hijri calendar. Both are derived from the arithmetic Islamic calendar
// (calc package). The civil holidays are proclaimed by Morocco's Ministry of
// Habous and Islamic Affairs and in some years land a day off from the
// arithmetic result; maMinistryOverrides pins those years to the proclaimed
// date. Years with no entry fall back to the calculation. See holidays#392.
func init() {
	RegisterMethod("eid_al_fitr", func(a MethodArgs) (time.Time, error) {
		return maFeastDate("eid_al_fitr", a.Year, maShawwal, 1)
	})
	RegisterMethod("eid_al_adha", func(a MethodArgs) (time.Time, error) {
		return maFeastDate("eid_al_adha", a.Year, maDhuAlHijjah, 10)
	})
}

// maFeastDate returns the Gregorian date of the given Hijri month/day
// occurrence in year, preferring a Ministry-proclaimed override when one is
// recorded. If neither an override nor a calculated occurrence exists for
// year (the arithmetic calendar's fixed +/-1-year search window can miss an
// occurrence near its edges, e.g. 1970-1973 for Dhu al-Hijjah 10, the same
// gap engine.trFeastDate documents for Turkey's sacrifice_feast), it returns
// the zero Time with a nil error: the same "no holiday this year" convention
// used by even_year_election_day in methods_us.go, not a hard failure.
// Ruby's CustomMethods::MA#feast likewise returns nil in this case and the
// gem treats that as no holiday, not an error.
func maFeastDate(method string, year, hijriMonth, hijriDay int) (time.Time, error) {
	if d, ok := maMinistryOverrides[maFeastKey{method, year}]; ok {
		return d, nil
	}
	d, ok := calc.HijriYearOccurrence(year, hijriMonth, hijriDay)
	if !ok {
		return time.Time{}, nil
	}
	return d, nil
}
