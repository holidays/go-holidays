package holidays

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/holidays/go-holidays/internal/engine"
	"github.com/holidays/go-holidays/internal/generator"
)

type loadedFile struct {
	key string
	rf  *generator.RegionFile
}

// RegisterMethod registers fn under name so a custom YAML rule can call it as
// `function: name(...)` or `observed: name(...)`. Register every method a file
// uses before calling LoadCustom on it.
//
// For a `function:` rule, fn returns the holiday's date for args.Year. Only the
// month and day of the returned date are kept: the rule's `function_modifier:`
// days are added first, then the year is forced to the year being resolved and
// the location to UTC. Returning the zero time.Time means the holiday does not
// occur that year.
//
// For an `observed:` rule, fn runs only when Options.Observed is true, and the
// date it returns replaces the holiday's date.
//
// fn receives only a MethodArgs. Arguments written in the YAML call, as in
// `my_method(year)`, must be syntactically valid but are not passed to fn.
//
// RegisterMethod panics if name is already registered, including any built-in
// method name, and a registered method cannot be removed. Register each name
// once, for example from an init function.
func RegisterMethod(name string, fn func(args MethodArgs) (time.Time, error)) {
	engine.RegisterMethod(name, engine.Method(fn))
}

// LoadCustom parses one or more holiday-definition YAML files (same schema as
// upstream holidays/definitions) and registers their rules at runtime,
// alongside the built-in rules. Query them by the codes in each rule's
// `regions:` list; like Options.Regions, those codes are lowercased and
// trimmed.
//
// Each file is registered under its basename without the extension, prefixed
// with "custom:". Only the basename matters: loading a path again replaces its
// prior load, and so does loading a different path with the same basename
// (/a/holidays.yaml and /b/holidays.yaml, or holidays.yaml and holidays.yml).
// Files with different basenames add rules without overwriting one another.
//
// The YAML's `methods:` block is parsed but its source is ignored. Every
// `function:` or `observed:` reference must name a method already registered
// via RegisterMethod, otherwise LoadCustom returns an error.
//
// LoadCustom is all-or-nothing: if any file cannot be read, parsed, or
// validated, it returns an error and registers none of them. It also returns an
// error when called with no paths. On success it calls ResetCache, since the
// rule set changed.
func LoadCustom(paths ...string) error {
	if len(paths) == 0 {
		return fmt.Errorf("holidays.LoadCustom: at least one path required")
	}
	loaded := make([]loadedFile, 0, len(paths))
	for _, p := range paths {
		lf, err := parseAndValidate(p)
		if err != nil {
			return err
		}
		loaded = append(loaded, lf)
	}
	// All files parsed and validated, so register them. We do the registration
	// in a second pass so a failure in any file leaves the registry untouched.
	for _, lf := range loaded {
		engine.RegisterCountry("custom:"+lf.key, lf.rf.Rules)
	}
	ResetCache()
	return nil
}

// UnloadCustom removes rules previously loaded by LoadCustom. The path is the
// same one passed to LoadCustom; only the basename matters, and a path that was
// never loaded is ignored. Methods registered with RegisterMethod stay
// registered. Intended for tests and long-running processes that need to drop
// reloadable rules. Calls ResetCache when it succeeds.
func UnloadCustom(paths ...string) {
	for _, p := range paths {
		base := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
		engine.UnregisterCountry("custom:" + base)
	}
	ResetCache()
}

// normalizeCustomRegionCodes lowercases and trims every region code on every
// rule in a freshly parsed custom file. The request path (normalizeRegions in
// holidays.go) does the same to Options.Regions, so without this a custom rule
// written with mixed-case or padded codes (e.g. "MyTeam") would be unreachable
// by any query. Built-in generated regions are already lowercase, so this only
// matters for LoadCustom. region_names: keys are normalized to match.
func normalizeCustomRegionCodes(rf *generator.RegionFile) {
	norm := func(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
	for i := range rf.Rules {
		for j := range rf.Rules[i].Regions {
			rf.Rules[i].Regions[j] = norm(rf.Rules[i].Regions[j])
		}
	}
	if len(rf.RegionNames) > 0 {
		normalized := make(map[string]string, len(rf.RegionNames))
		for code, name := range rf.RegionNames {
			normalized[norm(code)] = name
		}
		rf.RegionNames = normalized
	}
}

func parseAndValidate(path string) (loadedFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return loadedFile{}, fmt.Errorf("holidays.LoadCustom: read %s: %w", path, err)
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	if base == "" {
		return loadedFile{}, fmt.Errorf("holidays.LoadCustom: %s: cannot derive registry key from filename", path)
	}
	rf, err := generator.ParseRegionFile(base, data)
	if err != nil {
		return loadedFile{}, fmt.Errorf("holidays.LoadCustom: %s: %w", path, err)
	}
	normalizeCustomRegionCodes(rf)
	for _, r := range rf.Rules {
		if r.Function != "" && !engine.IsMethodRegistered(r.Function) {
			return loadedFile{}, fmt.Errorf("holidays.LoadCustom: %s: rule %q references unregistered method %q; call holidays.RegisterMethod first",
				path, r.Name, r.Function)
		}
		if r.Observed != "" && !engine.IsMethodRegistered(r.Observed) {
			return loadedFile{}, fmt.Errorf("holidays.LoadCustom: %s: rule %q references unregistered observed method %q",
				path, r.Name, r.Observed)
		}
	}
	return loadedFile{key: base, rf: rf}, nil
}
