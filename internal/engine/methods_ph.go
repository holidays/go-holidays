package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

func init() {
	// Philippines National Heroes Day: last Monday of August. This is the
	// correct, statutory rule. Upstream ph.yaml (holidays/definitions#345) has
	// an off-by-one that emits September 1 when August 31 is a Sunday; Go does
	// not reproduce that quirk. Behavior is locked by
	// TestPH_NationalHeroesDay_LastMondayOfAugust in pkg/ph_heroes_day_test.go.
	RegisterMethod("ph_heroes_day", func(a MethodArgs) (time.Time, error) {
		return calc.DayOfMonth(a.Year, time.August, -1, time.Monday), nil
	})
}
