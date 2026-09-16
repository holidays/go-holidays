package definition

// HolidayType distinguishes a formal, statutory holiday from an informal observance.
type HolidayType int

const (
	Formal HolidayType = iota
	Informal
)

// YearRangeKind identifies which shape of year restriction a YearRange applies:
// any year, until/from a bound year, a limited set of years, or a between span.
type YearRangeKind int

const (
	YearRangeAny YearRangeKind = iota
	YearRangeUntil
	YearRangeFrom
	YearRangeLimited
	YearRangeBetween
)

// YearRange restricts a HolidayRule to a subset of years, as determined by Kind
// and the Years it carries (their meaning depends on Kind).
type YearRange struct {
	Kind  YearRangeKind
	Years []int
}

// HolidayRule is one upstream holiday definition: its name, the regions it
// applies to, how its date is computed (a fixed month/day, an nth weekday, or a
// named function), and any observed-date or year-range restrictions.
type HolidayRule struct {
	Name             string
	Regions          []string
	Month            int
	Mday             int
	Wday             int
	Week             int
	Function         string
	FunctionArgs     []string
	FunctionModifier int
	Observed         string
	Type             HolidayType
	YearRanges       []YearRange
}

// HasWday reports whether r is computed from an nth-weekday-of-month rule.
func (r HolidayRule) HasWday() bool { return r.Week != 0 }

// HasMday reports whether r is computed from a fixed month/day.
func (r HolidayRule) HasMday() bool { return r.Mday != 0 }

// HasFunction reports whether r's date is computed by a named function.
func (r HolidayRule) HasFunction() bool { return r.Function != "" }

// HasObserved reports whether r has an observed-date adjustment function.
func (r HolidayRule) HasObserved() bool { return r.Observed != "" }

// AppliesIn reports whether r applies in the given year, honoring its
// YearRanges (an empty YearRanges means r applies in every year).
func (r HolidayRule) AppliesIn(year int) bool {
	if len(r.YearRanges) == 0 {
		return true
	}
	for _, yr := range r.YearRanges {
		if yr.Matches(year) {
			return true
		}
	}
	return false
}

// Matches reports whether year satisfies yr's Kind and Years restriction.
func (yr YearRange) Matches(year int) bool {
	switch yr.Kind {
	case YearRangeUntil:
		return len(yr.Years) == 1 && year <= yr.Years[0]
	case YearRangeFrom:
		return len(yr.Years) == 1 && year >= yr.Years[0]
	case YearRangeLimited:
		for _, y := range yr.Years {
			if y == year {
				return true
			}
		}
		return false
	case YearRangeBetween:
		return len(yr.Years) == 2 && year >= yr.Years[0] && year <= yr.Years[1]
	default:
		return true
	}
}
