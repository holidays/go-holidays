package engine

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"
)

func init() {
	// Tomb-Sweeping Day (清明节): falls on the Qingming solar term, April 4 or 5.
	RegisterMethod("cn_qingming", func(a MethodArgs) (time.Time, error) {
		return calc.Qingming(a.Year)
	})
}
