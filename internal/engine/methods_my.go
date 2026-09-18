package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

// Hijri month numbers used to locate Hari Raya Puasa (Eid al-Fitr) and Hari
// Raya Haji (Eid al-Adha).
const (
	myShawwal     = 10
	myDhuAlHijjah = 12
)

// myFeastKey identifies a gazette override entry.
type myFeastKey struct {
	method string
	year   int
}

// myGazetteOverrides holds gazetted first days that differ from the pure
// arithmetic Hijri calculation, keyed by (method, Gregorian year). Ported
// verbatim from Holidays::Definition::CustomMethods::MY::GAZETTE_OVERRIDES in
// the definitions repo (definitions PR #392, holidays/holidays#392). Years
// with no entry fall back to calc.HijriYearOccurrence.
var myGazetteOverrides = map[myFeastKey]time.Time{
	{"hari_raya_puasa", 2014}: time.Date(2014, time.July, 28, 0, 0, 0, 0, time.UTC),
	{"hari_raya_puasa", 2015}: time.Date(2015, time.July, 17, 0, 0, 0, 0, time.UTC),
	{"hari_raya_puasa", 2016}: time.Date(2016, time.July, 6, 0, 0, 0, 0, time.UTC),
	{"hari_raya_puasa", 2017}: time.Date(2017, time.June, 25, 0, 0, 0, 0, time.UTC),
	{"hari_raya_puasa", 2026}: time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC),

	{"hari_raya_haji", 2016}: time.Date(2016, time.September, 12, 0, 0, 0, 0, time.UTC),
	{"hari_raya_haji", 2017}: time.Date(2017, time.September, 1, 0, 0, 0, 0, time.UTC),
	{"hari_raya_haji", 2019}: time.Date(2019, time.August, 11, 0, 0, 0, 0, time.UTC),
}

// Hari Raya Puasa (Eid al-Fitr) and Hari Raya Haji (Eid al-Adha) are tied to
// the Hijri calendar. Both are derived from the arithmetic Islamic calendar
// (calc package). The civil holidays are fixed by the Yang di-Pertuan Agong
// (Conference of Rulers) on JAKIM's moon-sighting advice and in some years
// have landed a day before the arithmetic result; myGazetteOverrides pins
// those years to the gazetted date. Years with no entry fall back to the
// calculation. See holidays/holidays#392.
func init() {
	RegisterMethod("hari_raya_puasa", func(a MethodArgs) (time.Time, error) {
		return myFeastDate("hari_raya_puasa", a.Year, myShawwal, 1)
	})
	RegisterMethod("hari_raya_haji", func(a MethodArgs) (time.Time, error) {
		return myFeastDate("hari_raya_haji", a.Year, myDhuAlHijjah, 10)
	})
}

// myFeastDate returns the Gregorian date of the given Hijri month/day
// occurrence in year, preferring a gazette-proclaimed override when one is
// recorded. If neither an override nor a calculated occurrence exists for
// year (the arithmetic calendar's fixed +/-1-year search window can miss an
// occurrence near its edges, e.g. 1970-1973 for Dhu al-Hijjah 10 - the same
// gap tr's sacrifice_feast has, since it is the same Hijri month/day), it
// returns the zero Time with a nil error: the same "no holiday this year"
// convention used by trFeastDate in methods_tr.go. Ruby's
// CustomMethods::MY#feast likewise returns nil in this case and the gem
// treats that as no holiday, not an error.
func myFeastDate(method string, year, hijriMonth, hijriDay int) (time.Time, error) {
	if d, ok := myGazetteOverrides[myFeastKey{method, year}]; ok {
		return d, nil
	}
	d, ok := calc.HijriYearOccurrence(year, hijriMonth, hijriDay)
	if !ok {
		return time.Time{}, nil
	}
	return d, nil
}
