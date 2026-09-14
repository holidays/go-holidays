package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

func init() {
	// Seollal eve (설날 연휴): the day before Seollal, which is the first day of
	// the first lunar month.
	RegisterMethod("kr_seollal_eve", func(a MethodArgs) (time.Time, error) {
		seollal, err := calc.LunarToSolar(a.Year, 1, 1, a.Region)
		if err != nil {
			return time.Time{}, err
		}
		return seollal.AddDate(0, 0, -1), nil
	})
}
