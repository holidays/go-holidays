package engine

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/holidays/go-holidays/internal/definition"
)

// MethodArgs is the uniform invocation payload for every registered method.
// The resolver populates every applicable field; each method picks what it needs.
type MethodArgs struct {
	Year   int
	Month  int
	Day    int
	Date   time.Time
	Region string
}

// Method computes or transforms a date.
type Method func(args MethodArgs) (time.Time, error)

var (
	methodMu sync.RWMutex
	methods  = map[string]Method{}

	regionMu  sync.RWMutex
	countries = map[string][]definition.HolidayRule{} // keyed by country code (us, gb, ...)

	regionNameMu sync.RWMutex
	regionNames  = map[string]string{} // region code -> display name, flat/global

	// flatRulesCache is the concatenation of every registered country's rules,
	// rebuilt lazily on first use after (un)registration invalidates it. See
	// rulesFor: every ResolveYear call used to rebuild this ~187k-entry slice
	// from scratch, which dominated Go-side sweep cost.
	flatRulesCache []definition.HolidayRule
	flatRulesValid bool

	// afterInvalidRead is a test-only hook; see rulesFor.
	afterInvalidRead func()
)

// RegisterMethod registers a named method. Panics if name is already registered.
func RegisterMethod(name string, fn Method) {
	methodMu.Lock()
	defer methodMu.Unlock()
	if _, exists := methods[name]; exists {
		panic(fmt.Sprintf("engine: method %q already registered", name))
	}
	methods[name] = fn
}

// LookupMethod returns the registered method for name.
func LookupMethod(name string) (Method, bool) {
	methodMu.RLock()
	defer methodMu.RUnlock()
	fn, ok := methods[name]
	return fn, ok
}

// IsMethodRegistered reports whether name has been registered.
func IsMethodRegistered(name string) bool {
	_, ok := LookupMethod(name)
	return ok
}

// RegisterCountry installs the rules for a country YAML (e.g. "us" from us.yaml).
// Subregions like us_ga are addressable via these same rules — each rule carries
// its own Regions slice.
func RegisterCountry(country string, rules []definition.HolidayRule) {
	regionMu.Lock()
	defer regionMu.Unlock()
	countries[country] = rules
	flatRulesValid = false
}

// UnregisterCountry removes a registered country key. Used by LoadCustom
// callers (typically tests) to drop runtime-loaded rules.
func UnregisterCountry(country string) {
	regionMu.Lock()
	defer regionMu.Unlock()
	delete(countries, country)
	flatRulesValid = false
}

// RegisterRegionNames installs the region code to display name map for a
// country YAML (e.g. "au" from au.yaml). Registrations merge into one flat,
// global map; a code registered by more than one file is last-write-wins.
// Current upstream data has no conflicting values for a duplicate code (only
// exact duplicates, e.g. "us" registered identically by both us.yaml and
// northamericainformal.yaml), so this is safe today, but a future upstream
// edit that introduces a genuine conflict would silently pick one value
// rather than erroring.
func RegisterRegionNames(country string, names map[string]string) {
	regionNameMu.Lock()
	defer regionNameMu.Unlock()
	for code, name := range names {
		regionNames[code] = name
	}
}

// RegionName returns the display name registered for region, and whether it
// was found.
func RegionName(region string) (string, bool) {
	regionNameMu.RLock()
	defer regionNameMu.RUnlock()
	name, ok := regionNames[region]
	return name, ok
}

// RegionNames returns a copy of every registered region code mapped to its
// display name. A copy is returned so callers cannot mutate internal state.
func RegionNames() map[string]string {
	regionNameMu.RLock()
	defer regionNameMu.RUnlock()
	out := make(map[string]string, len(regionNames))
	for k, v := range regionNames {
		out[k] = v
	}
	return out
}

// AvailableRegions returns every region code mentioned across all registered rules,
// sorted lexicographically.
func AvailableRegions() []string {
	regionMu.RLock()
	defer regionMu.RUnlock()
	seen := map[string]struct{}{}
	for _, rules := range countries {
		for _, r := range rules {
			for _, reg := range r.Regions {
				seen[reg] = struct{}{}
			}
		}
	}
	out := make([]string, 0, len(seen))
	for r := range seen {
		out = append(out, r)
	}
	sort.Strings(out)
	return out
}

// rulesFor returns the rules to scan given requested regions. Since registered
// country keys do not always align with region prefixes (e.g. be_fr.yaml
// registers under "be_fr", not "be"), we walk every registered country's
// rules and let ruleMatchesRequested decide. The requested regions do not
// narrow this slice (filtering happens per-rule in ResolveYear via
// ruleMatchesRequested), so the flat concatenation of every registered
// country's rules is identical across calls and worth caching: rebuilding it
// on every ResolveYear call re-walked and re-appended ~187k rule entries.
//
// The cache is invalidated by RegisterCountry/UnregisterCountry (LoadCustom
// calls these at runtime), so a stale slice is never served across a
// registration change. Double-checked locking: the common case (cache
// already valid) only takes the read lock.
func rulesFor(requested []string) []definition.HolidayRule {
	regionMu.RLock()
	if flatRulesValid {
		out := flatRulesCache
		regionMu.RUnlock()
		return out
	}
	regionMu.RUnlock()

	// Test-only seam: lets tests deterministically reproduce the race the
	// double check below guards against (another goroutine rebuilding the
	// cache between our read unlock and our write lock). Nil in production.
	if afterInvalidRead != nil {
		afterInvalidRead()
	}

	regionMu.Lock()
	defer regionMu.Unlock()
	if flatRulesValid {
		return flatRulesCache
	}
	var out []definition.HolidayRule
	for _, rs := range countries {
		out = append(out, rs...)
	}
	flatRulesCache = out
	flatRulesValid = true
	return out
}

// rulesForCountry returns the rules registered for the named country file.
// Method implementations call this at resolve time, after all init() functions
// have run, to inspect peer rules (e.g. JP substitute-holiday logic needs to
// know which days are already holidays).
func rulesForCountry(country string) []definition.HolidayRule {
	regionMu.RLock()
	defer regionMu.RUnlock()
	return countries[country]
}

// ruleMatchesRequested reports whether a rule applies to any of the requested regions.
// Matching rules:
//   - Exact equality: rule region == requested region.
//   - Parent: rule region "us" applies when the user requests "us_ga".
//   - Wildcard: a requested region ending with "_" always collapses to its
//     country segment (the part before the first "_"), regardless of how
//     many segments follow. "gb_" and the multi-segment "gb_eng_" both match
//     any rule region that equals that country segment ("gb") or starts with
//     the country segment plus "_" ("gb_sct", "gb_wls", "gb_eng"): a
//     multi-segment wildcard is not scoped to its own branch, it behaves
//     identically to the bare country wildcard.
func ruleMatchesRequested(rule definition.HolidayRule, requested []string) bool {
	if len(requested) == 0 {
		return true
	}
	for _, req := range requested {
		if strings.HasSuffix(req, "_") {
			bare := strings.SplitN(req, "_", 2)[0]
			for _, rr := range rule.Regions {
				if rr == bare || strings.HasPrefix(rr, bare+"_") {
					return true
				}
			}
			continue
		}
		for _, rr := range rule.Regions {
			if rr == req || isParentOf(rr, req) {
				return true
			}
		}
	}
	return false
}

func isParentOf(parent, child string) bool {
	return len(child) > len(parent) && strings.HasPrefix(child, parent+"_")
}
