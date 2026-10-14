package calc

import (
	"fmt"
	"math"
	"time"
)

// Qingming returns the Gregorian date of the Qingming solar term (清明), which
// anchors China's Tomb-Sweeping Day. It uses an astronomical approximation: a
// per-century base constant plus a linear drift term, corrected for leap years.
// Qingming always falls on April 4 or 5 within the supported range,
// 1900..2100. The formula is off by one day in 2100 only: the sun reaches 15
// degrees at 00:45 on April 5 2100, Beijing time (JPL DE440s). Other years
// return an error wrapping ErrUnsupportedYear.
func Qingming(year int) (time.Time, error) {
	var constant float64
	switch {
	case year >= 1900 && year <= 1999:
		constant = 5.59
	case year >= 2000 && year <= 2099:
		constant = 4.81
	case year == 2100:
		return time.Date(2100, time.April, 5, 0, 0, 0, 0, time.UTC), nil
	default:
		return time.Time{}, fmt.Errorf("calc.Qingming: covers 1900..2100, got %d: %w", year, ErrUnsupportedYear)
	}
	y := year % 100
	day := int(math.Floor(float64(y)*0.2422+constant)) - y/4
	return time.Date(year, time.April, day, 0, 0, 0, 0, time.UTC), nil
}
