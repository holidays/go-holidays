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

// MethodArgs re-exports engine.MethodArgs so callers using RegisterMethod do
// not need to import internal/engine.
type MethodArgs = engine.MethodArgs
