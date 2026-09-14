package holidays

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/holidays/go-holidays/internal/engine"
)

var (
	cacheMu         sync.RWMutex
	cacheStore_data = map[cacheKey]cacheEntry{}
)

type cacheKey struct {
	regions  string
	informal bool
	observed bool
}

type cacheEntry struct {
	from, to time.Time
	byDay    map[string][]Holiday
}

// CacheBetween pre-computes and stores every holiday in [from, to] for the given
// options. Subsequent calls to On/Between with the same options and a range
// fully contained in [from, to] will be answered from the cache.
func CacheBetween(from, to time.Time, opts Options) error {
	opts.Regions = normalizeRegions(opts.Regions)
	hs, err := computeBetween(from, to, opts)
	if err != nil {
		return err
	}
	cacheStore(startOfDay(from), startOfDay(to), opts, hs)
	return nil
}

// ResetCache clears every cached range. Intended for tests and long-running
// processes that need to drop stale entries.
func ResetCache() {
	cacheMu.Lock()
	cacheStore_data = map[cacheKey]cacheEntry{}
	cacheMu.Unlock()
}

func optionsKey(opts Options) cacheKey {
	rs := append([]string(nil), opts.Regions...)
	sort.Strings(rs)
	return cacheKey{
		regions:  strings.Join(rs, ","),
		informal: opts.Informal,
		observed: opts.Observed,
	}
}

func dayKey(t time.Time) string {
	return t.Format("2006-01-02")
}

// cacheFind returns cached holidays in [start, end] for opts iff a previously
// cached range fully contains [start, end] for the same options.
func cacheFind(start, end time.Time, opts Options) ([]Holiday, bool) {
	cacheMu.RLock()
	entry, ok := cacheStore_data[optionsKey(opts)]
	cacheMu.RUnlock()
	if !ok {
		return nil, false
	}
	if start.Before(entry.from) || end.After(entry.to) {
		return nil, false
	}
	var out []Holiday
	for _, list := range entry.byDay {
		for _, h := range list {
			if !h.Date.Before(start) && !h.Date.After(end) {
				out = append(out, h)
			}
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if !out[i].Date.Equal(out[j].Date) {
			return out[i].Date.Before(out[j].Date)
		}
		return out[i].Name < out[j].Name
	})
	return out, true
}

func cacheStore(from, to time.Time, opts Options, holidays []Holiday) {
	byDay := map[string][]Holiday{}
	for _, h := range holidays {
		k := dayKey(h.Date)
		byDay[k] = append(byDay[k], h)
	}
	cacheMu.Lock()
	cacheStore_data[optionsKey(opts)] = cacheEntry{from: from, to: to, byDay: byDay}
	cacheMu.Unlock()
}

// computeBetween is the cache-bypassing implementation of Between. Used by
// Between (on cache miss) and by CacheBetween to populate the cache.
func computeBetween(start, end time.Time, opts Options) ([]Holiday, error) {
	if end.Before(start) {
		return nil, fmt.Errorf("holidays.Between: end %s is before start %s",
			end.Format("2006-01-02"), start.Format("2006-01-02"))
	}
	var out []Holiday
	// Resolve one year past end.Year() as well, then clip to [start, end]: a
	// next-year holiday whose observed date shifts back into range (for example
	// New Year's Day Jan 1 observed on the preceding Dec 31) lives in the
	// following year's ResolveYear. The inRange clip keeps only in-window dates,
	// and a given holiday instance appears in only one year's ResolveYear, so
	// there is no duplication.
	for y := start.Year(); y <= end.Year()+1; y++ {
		resolved, err := engine.ResolveYear(y, engine.ResolveOptions{
			Regions:  opts.Regions,
			Informal: opts.Informal,
			Observed: opts.Observed,
		})
		if err != nil {
			return nil, err
		}
		for _, r := range resolved {
			if inRange(r.Date, start, end) {
				out = append(out, Holiday(r))
			}
		}
	}
	return out, nil
}
