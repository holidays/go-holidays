package holidays

import (
	"time"

	"github.com/holidays/go-holidays/internal/engine"
)

// Holiday is one resolved holiday occurrence: a date, its name, the regions it
// applies to, and whether it is an informal observance.
type Holiday struct {
	Date    time.Time
	Name    string
	Regions []string
	// Informal reports whether this holiday is an informal observance (for
	// example Valentine's Day or Halloween) rather than a formal, statutory
	// holiday. It mirrors the rule's upstream type and only distinguishes
	// results when Options.Informal was set to true on the request that
	// produced them: with Options.Informal false, informal holidays are
	// filtered out during resolution and never appear, so every returned
	// Holiday has Informal == false.
	Informal bool
}

// Options controls a holiday lookup. An empty Regions slice means "all registered regions".
type Options struct {
	Regions  []string
	Informal bool
	Observed bool
}

// MethodArgs is the input to a method passed to RegisterMethod. It re-exports
// engine.MethodArgs so callers do not need to import internal/engine. Its
// fields are Year, Month, Day (ints), Date (time.Time) and Region (string), and
// what they hold depends on how the method is called.
//
// For a `function:` rule, Year is the year being resolved; Month and Day are
// the rule's month and mday (Day is 0 when the rule uses wday instead); Date is
// the date computed from the rule's month and its mday or wday/week, or the
// zero time when the rule has neither; Region is the first region listed on
// the rule.
//
// For an `observed:` rule, Date is the holiday's date and Year, Month and Day
// are its parts; Region is the first region in Options.Regions, or "" when the
// request named none.
type MethodArgs = engine.MethodArgs
