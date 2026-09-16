package holidays

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/holidays/go-holidays/internal/engine"

	_ "github.com/holidays/go-holidays/internal/definitions"
)

// nextHolidaysMaxForwardYears bounds the forward scan in NextHolidays so a
// region with too few holidays cannot loop unbounded; the loop stops once it
// has collected `count` holidays or has scanned this many years past `from`.
const nextHolidaysMaxForwardYears = 100

// On returns every holiday matching the given options that falls on date.
func On(date time.Time, opts Options) ([]Holiday, error) {
	return Between(date, date, opts)
}

// Between returns every holiday matching the given options whose date falls in
// [start, end] (inclusive on both ends, compared by calendar day).
func Between(start, end time.Time, opts Options) ([]Holiday, error) {
	opts.Regions = normalizeRegions(opts.Regions)
	if end.Before(start) {
		return nil, fmt.Errorf("holidays.Between: end %s is before start %s",
			end.Format("2006-01-02"), start.Format("2006-01-02"))
	}
	if hs, ok := cacheFind(startOfDay(start), startOfDay(end), opts); ok {
		return hs, nil
	}
	return computeBetween(start, end, opts)
}

// YearHolidays returns every holiday matching the given options in the given year.
func YearHolidays(year int, opts Options) ([]Holiday, error) {
	opts.Regions = normalizeRegions(opts.Regions)
	resolved, err := engine.ResolveYear(year, engine.ResolveOptions{
		Regions:  opts.Regions,
		Informal: opts.Informal,
		Observed: opts.Observed,
	})
	if err != nil {
		return nil, err
	}
	out := make([]Holiday, len(resolved))
	for i, r := range resolved {
		out[i] = Holiday(r)
	}
	return out, nil
}

// YearHolidaysFrom returns every holiday matching the given options from `from`
// (truncated to its UTC calendar day) through Dec 31 of `from`'s year, sorted by
// date ascending. It clips a 12-month forward window to Dec 31, so the result
// includes next-year holidays whose observed date shifts back on or before Dec 31
// of `from`'s year (for example New Year's Day observed on Dec 31).
func YearHolidaysFrom(from time.Time, opts Options) ([]Holiday, error) {
	opts.Regions = normalizeRegions(opts.Regions)
	var (
		fromDay     = startOfDay(from)
		upper       = time.Date(fromDay.Year(), 12, 31, 0, 0, 0, 0, fromDay.Location())
		resolveOpts = engine.ResolveOptions{
			Regions:  opts.Regions,
			Informal: opts.Informal,
			Observed: opts.Observed,
		}
		out []Holiday
	)
	for i, year := range []int{fromDay.Year(), fromDay.Year() + 1} {
		resolved, err := engine.ResolveYear(year, resolveOpts)
		if err != nil {
			// The primary year must resolve. The following year is only a
			// look-ahead to pull boundary-adjacent holidays whose observed date
			// shifts back into [fromDay, Dec 31]
			if i == 0 {
				return nil, err
			}
			break
		}
		for _, r := range resolved {
			if r.Date.Before(fromDay) || r.Date.After(upper) {
				continue
			}
			out = append(out, Holiday(r))
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Date.Before(out[j].Date)
	})
	return out, nil
}

// NextHolidays returns the next `count` holidays on or after `from`, sorted by
// date ascending. It keeps resolving forward, year by year, accumulating
// holidays with date >= `from` until at least `count` are gathered, then
// truncates to `count`.
func NextHolidays(from time.Time, count int, opts Options) ([]Holiday, error) {
	if count <= 0 {
		return nil, fmt.Errorf("holidays.NextHolidays: count must be positive, got %d", count)
	}
	opts.Regions = normalizeRegions(opts.Regions)
	var (
		fromDay     = startOfDay(from)
		resolveOpts = engine.ResolveOptions{
			Regions:  opts.Regions,
			Informal: opts.Informal,
			Observed: opts.Observed,
		}
		collected []Holiday
		startYear = fromDay.Year()
	)
	for offset := 0; offset <= nextHolidaysMaxForwardYears; offset++ {
		resolved, err := engine.ResolveYear(startYear+offset, resolveOpts)
		if err != nil {
			return nil, err
		}
		for _, r := range resolved {
			if r.Date.Before(fromDay) {
				continue
			}
			collected = append(collected, Holiday(r))
		}
		if len(collected) >= count {
			break
		}
	}
	sort.SliceStable(collected, func(i, j int) bool {
		return collected[i].Date.Before(collected[j].Date)
	})
	if len(collected) > count {
		collected = collected[:count]
	}
	return collected, nil
}

// AnyHolidaysDuringWorkWeek reports whether any holiday matching opts falls
// during the Mon-Fri work week containing `date`. For a Saturday input, the
// work week is the preceding Mon-Fri; for a Sunday input, the following.
func AnyHolidaysDuringWorkWeek(date time.Time, opts Options) (bool, error) {
	var (
		d      = startOfDay(date)
		wday   = int(d.Weekday())
		monday = d.AddDate(0, 0, -(wday - 1))
		friday = d.AddDate(0, 0, 5-wday)
	)
	hs, err := Between(monday, friday, opts)
	if err != nil {
		return false, err
	}
	return len(hs) > 0, nil
}

// AvailableRegions returns every region code registered, sorted lexicographically.
func AvailableRegions() []string {
	return engine.AvailableRegions()
}

// RegionName returns the display name for a region code, and whether it is
// registered. Unregistered codes return ("", false).
func RegionName(region string) (string, bool) {
	return engine.RegionName(region)
}

// RegionNames returns every registered region code mapped to its display name.
func RegionNames() map[string]string {
	return engine.RegionNames()
}

// normalizeRegions cleans up the requested region codes so that "US", " us ",
// and "us" all behave identically: each code is trimmed of surrounding
// whitespace and lowercased, and codes that are empty after trimming are
// dropped. A nil or empty slice (meaning "all regions") is returned unchanged.
// This runs only on the request path (Options.Regions); region registration
// and AvailableRegions are untouched. A genuinely unknown code still yields an
// empty result rather than an error, as everywhere else in the package.
func normalizeRegions(regions []string) []string {
	if len(regions) == 0 {
		return regions
	}
	out := make([]string, 0, len(regions))
	for _, r := range regions {
		if r = strings.ToLower(strings.TrimSpace(r)); r != "" {
			out = append(out, r)
		}
	}
	return out
}

func inRange(d, start, end time.Time) bool {
	if d.Before(startOfDay(start)) {
		return false
	}
	if d.After(startOfDay(end)) {
		return false
	}
	return true
}

func startOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}
