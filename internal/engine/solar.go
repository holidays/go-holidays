package engine

import (
	"fmt"
	"math"
	"time"
)

// solarTermRange pairs an inclusive year range with the base constant the
// solar-term formula uses for those years.
type solarTermRange struct {
	lo, hi int
	base   float64
}

// solarTermBase is the set of base constants for one equinox or solstice.
type solarTermBase struct {
	ranges []solarTermRange
}

// solarTermDay computes the day in month of an equinox or solstice for year
// using the common approximation floor(base + 0.242194*(y-1980) - floor((y-1980)/4)).
// It errors when year falls outside every range in sb.
func solarTermDay(year int, month time.Month, sb solarTermBase) (time.Time, error) {
	var (
		base  float64
		found bool
	)
	for _, r := range sb.ranges {
		if year >= r.lo && year <= r.hi {
			base = r.base
			found = true
			break
		}
	}
	if !found {
		return time.Time{}, fmt.Errorf("solar term: year %d out of supported range", year)
	}
	diff := float64(year - 1980)
	day := math.Floor(base + 0.242194*diff - math.Floor(diff/4))
	return time.Date(year, month, int(day), 0, 0, 0, 0, time.UTC), nil
}
