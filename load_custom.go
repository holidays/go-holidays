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

// RegisterMethod registers a named method usable from a YAML rule's
// `function:` or `observed:` field. Must be called before LoadCustom for any
// custom YAML that references the name. Wraps engine.RegisterMethod.
func RegisterMethod(name string, fn func(args MethodArgs) (time.Time, error)) {
	engine.RegisterMethod(name, engine.Method(fn))
}

// LoadCustom parses one or more holiday-definition YAML files (same schema as
// upstream holidays/definitions) and registers their rules at runtime,
// alongside the built-in rules. Each file's basename (sans extension) becomes
// its registry key prefixed with "custom:", so re-loading the same path
// replaces its prior load; distinct paths add rules without overwriting one
// another.
//
// Any `function:` or `observed:` reference in user YAML must point to a method
// already registered via RegisterMethod — we cannot interpret Ruby method
// bodies. The YAML's `methods:` block is parsed but its Ruby source is ignored.
//
// LoadCustom calls ResetCache when it succeeds, since the rule set changed.
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
	// All files parsed and validated — register them. We do the registration
	// in a second pass so a failure in any file leaves the registry untouched.
	for _, lf := range loaded {
		engine.RegisterCountry("custom:"+lf.key, lf.rf.Rules)
	}
	ResetCache()
	return nil
}

// UnloadCustom removes rules previously loaded by LoadCustom. The path is the
// same one passed to LoadCustom; only the basename matters. Intended for tests
// and long-running processes that need to drop reloadable rules. Calls
// ResetCache when it succeeds.
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
