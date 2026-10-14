package calc

import (
	"errors"
	"fmt"
	"time"
)

// ErrUnsupportedYear is wrapped by errors for a year outside the range a
// calculation's tables or formula support, e.g. a lunar year past the end of
// the lunar tables. Test for it with errors.Is.
var ErrUnsupportedYear = errors.New("year outside supported range")

// firstLunarYear and lastLunarYear bound the lunar tables in lunar_data.go.
const (
	firstLunarYear = 1900
	lastLunarYear  = 2100
)

// solarStartDate is the Gregorian date corresponding to lunar 1900-01-01
// (Lunar New Year, January 31 1900). All conversions are deltas off this anchor.
var solarStartDate = time.Date(1900, time.January, 31, 0, 0, 0, 0, time.UTC)

// LunarToSolar converts a lunar calendar date (year/month/day in the region's
// lunar calendar) to its Gregorian equivalent. The lookup tables cover cn, hk,
// kr, sg, and vn for lunar years 1900..2100 (cn, hk and sg share the Chinese
// table); other years return an error wrapping ErrUnsupportedYear.
func LunarToSolar(year, month, day int, region string) (time.Time, error) {
	info := calendarYearInfoForRegion(region)
	if info == nil {
		return time.Time{}, fmt.Errorf("calc.LunarToSolar: no lunar table for region %q", region)
	}
	if year < firstLunarYear || year > lastLunarYear {
		return time.Time{}, fmt.Errorf("calc.LunarToSolar: lunar calendar data for %q covers lunar years %d..%d, got %d: %w",
			region, firstLunarYear, lastLunarYear, year, ErrUnsupportedYear)
	}
	yearDiff := year - firstLunarYear
	if month < 1 || month > 12 {
		return time.Time{}, fmt.Errorf("calc.LunarToSolar: invalid lunar month %d", month)
	}

	days := 0
	for i := 0; i < yearDiff; i++ {
		days += info[i][0]
	}
	for i := 0; i < month-1; i++ {
		days += lunarDaysForMonthType[info[yearDiff][i+1]][0]
	}
	days += day - 1
	return solarStartDate.AddDate(0, 0, days), nil
}
