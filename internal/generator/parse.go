package generator

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/holidays/go-holidays/internal/definition"
)

// RegionFile is the in-memory result of parsing one definitions YAML.
type RegionFile struct {
	Country     string // base filename without .yaml
	Rules       []definition.HolidayRule
	Methods     map[string]MethodSpec // local custom methods (validation only)
	Tests       []TestSpec
	RegionNames map[string]string // region code -> display name, from region_names:
}

// MethodSpec captures a custom method's declared signature. The Ruby body is
// retained for diagnostics only — we do not interpret it.
type MethodSpec struct {
	Name      string
	Arguments []string
	Ruby      string
}

// TestSpec is one test entry from the YAML's tests: block.
type TestSpec struct {
	Dates    []string
	Regions  []string
	Options  []string
	Expected ExpectedSpec
}

// ExpectedSpec mirrors the `expect:` sub-block. Holiday is nil if the YAML
// did not specify it (only `name:` was given).
type ExpectedSpec struct {
	Name    string
	Holiday *bool
}

// MethodCall is the parsed form of a `function:` or `observed:` expression.
type MethodCall struct {
	Name string
	Args []string
}

type rawRoot struct {
	Months      map[int][]rawRule    `yaml:"months"`
	Methods     map[string]rawMethod `yaml:"methods"`
	Tests       []rawTest            `yaml:"tests"`
	RegionNames map[string]string    `yaml:"region_names"`
}

type rawRule struct {
	Name             string         `yaml:"name"`
	Regions          []string       `yaml:"regions"`
	Mday             int            `yaml:"mday"`
	Wday             *int           `yaml:"wday"`
	Week             int            `yaml:"week"`
	Function         string         `yaml:"function"`
	FunctionModifier int            `yaml:"function_modifier"`
	Observed         string         `yaml:"observed"`
	Type             string         `yaml:"type"`
	YearRanges       *rawYearRanges `yaml:"year_ranges"`
}

type rawYearRanges struct {
	Until   *int            `yaml:"until"`
	From    *int            `yaml:"from"`
	Limited []int           `yaml:"limited"`
	Between *rawBetweenSpan `yaml:"between"`
}

type rawBetweenSpan struct {
	Start int `yaml:"start"`
	End   int `yaml:"end"`
}

type rawMethod struct {
	Arguments string `yaml:"arguments"`
	Ruby      string `yaml:"ruby"`
}

type rawTest struct {
	Given  rawTestGiven  `yaml:"given"`
	Expect rawTestExpect `yaml:"expect"`
}

type rawTestGiven struct {
	Date    yaml.Node `yaml:"date"` // string or sequence of strings
	Regions []string  `yaml:"regions"`
	Options yaml.Node `yaml:"options"` // string or sequence of strings
}

type rawTestExpect struct {
	Name    string `yaml:"name"`
	Holiday *bool  `yaml:"holiday"`
	Holyday *bool  `yaml:"holyday"` // typo in some upstream YAMLs; treat as alias
}

// ParseRegionFile parses one definitions YAML.
func ParseRegionFile(country string, data []byte) (*RegionFile, error) {
	var (
		root rawRoot
		dec  = yaml.NewDecoder(bytes.NewReader(data))
	)
	dec.KnownFields(true)
	if err := dec.Decode(&root); err != nil {
		return nil, fmt.Errorf("parse %s.yaml: %w", country, err)
	}

	rf := &RegionFile{Country: country, Methods: map[string]MethodSpec{}, RegionNames: root.RegionNames}

	months := make([]int, 0, len(root.Months))
	for m := range root.Months {
		months = append(months, m)
	}
	sort.Ints(months)
	for _, month := range months {
		for i, rr := range root.Months[month] {
			rule, err := convertRule(country, month, rr)
			if err != nil {
				return nil, fmt.Errorf("%s.yaml month %d entry %d: %w", country, month, i, err)
			}
			rf.Rules = append(rf.Rules, rule)
		}
	}

	for name, m := range root.Methods {
		rf.Methods[name] = MethodSpec{
			Name:      name,
			Arguments: splitArgs(m.Arguments),
			Ruby:      m.Ruby,
		}
	}

	for i, rt := range root.Tests {
		ts, skip, err := convertTest(country, i, rt)
		if err != nil {
			return nil, fmt.Errorf("%s.yaml test %d: %w", country, i, err)
		}
		if skip {
			continue
		}
		rf.Tests = append(rf.Tests, ts)
	}

	return rf, nil
}

func convertRule(country string, month int, rr rawRule) (definition.HolidayRule, error) {
	if rr.Name == "" {
		return definition.HolidayRule{}, fmt.Errorf("missing name")
	}
	if len(rr.Regions) == 0 {
		return definition.HolidayRule{}, fmt.Errorf("missing regions")
	}
	rule := definition.HolidayRule{
		Name:             rr.Name,
		Regions:          rr.Regions,
		Month:            month,
		Mday:             rr.Mday,
		Wday:             -1,
		Week:             rr.Week,
		FunctionModifier: rr.FunctionModifier,
	}
	if rr.Wday != nil {
		rule.Wday = *rr.Wday
	}
	switch strings.ToLower(rr.Type) {
	case "", "formal":
		rule.Type = definition.Formal
	case "informal":
		rule.Type = definition.Informal
	default:
		return definition.HolidayRule{}, fmt.Errorf("unknown type %q", rr.Type)
	}
	if rr.Function != "" {
		call, err := ParseCall(rr.Function)
		if err != nil {
			return definition.HolidayRule{}, fmt.Errorf("function %q: %w", rr.Function, err)
		}
		rule.Function = call.Name
		rule.FunctionArgs = call.Args
	}
	if rr.Observed != "" {
		call, err := ParseCall(rr.Observed)
		if err != nil {
			return definition.HolidayRule{}, fmt.Errorf("observed %q: %w", rr.Observed, err)
		}
		rule.Observed = call.Name
	}
	if rr.YearRanges != nil {
		yr, err := convertYearRange(*rr.YearRanges)
		if err != nil {
			return definition.HolidayRule{}, err
		}
		rule.YearRanges = []definition.YearRange{yr}
	}
	return rule, nil
}

func convertYearRange(rr rawYearRanges) (definition.YearRange, error) {
	count := 0
	if rr.Until != nil {
		count++
	}
	if rr.From != nil {
		count++
	}
	if rr.Limited != nil {
		count++
	}
	if rr.Between != nil {
		count++
	}
	if count != 1 {
		return definition.YearRange{}, fmt.Errorf("year_ranges must have exactly one of until/from/limited/between, got %d", count)
	}
	// The count != 1 guard above already ensures exactly one of these four
	// fields is set, so the last case (Between) is reached via default rather
	// than repeating the nil check on genuinely unreachable dead code.
	switch {
	case rr.Until != nil:
		return definition.YearRange{Kind: definition.YearRangeUntil, Years: []int{*rr.Until}}, nil
	case rr.From != nil:
		return definition.YearRange{Kind: definition.YearRangeFrom, Years: []int{*rr.From}}, nil
	case rr.Limited != nil:
		return definition.YearRange{Kind: definition.YearRangeLimited, Years: rr.Limited}, nil
	default:
		return definition.YearRange{Kind: definition.YearRangeBetween, Years: []int{rr.Between.Start, rr.Between.End}}, nil
	}
}

// convertTest returns (spec, skip, err). If skip is true the caller should drop
// this test entirely (e.g. its dates were all malformed in upstream YAML).
func convertTest(country string, idx int, rt rawTest) (TestSpec, bool, error) {
	// decodeDateNode never returns a nil error alongside an empty slice, so
	// there is no separate empty check to make here.
	dates, err := decodeDateNode(rt.Given.Date)
	if err != nil {
		return TestSpec{}, false, fmt.Errorf("given.date: %w", err)
	}
	valid, skipped := filterParseableDates(dates)
	if len(skipped) > 0 {
		fmt.Fprintf(os.Stderr, "  warn: %s.yaml test %d: skipping unparseable dates %v\n", country, idx, skipped)
	}
	if len(valid) == 0 {
		return TestSpec{}, true, nil
	}
	options, err := decodeStringOrList(rt.Given.Options)
	if err != nil {
		return TestSpec{}, false, fmt.Errorf("given.options: %w", err)
	}
	holiday := rt.Expect.Holiday
	if holiday == nil {
		holiday = rt.Expect.Holyday
	}
	if rt.Expect.Name == "" && holiday == nil {
		return TestSpec{}, false, fmt.Errorf("expect must specify name or holiday")
	}
	return TestSpec{
		Dates:   valid,
		Regions: rt.Given.Regions,
		Options: options,
		Expected: ExpectedSpec{
			Name:    rt.Expect.Name,
			Holiday: holiday,
		},
	}, false, nil
}

func filterParseableDates(in []string) (valid, skipped []string) {
	for _, s := range in {
		if _, err := time.Parse("2006-1-2", s); err != nil {
			skipped = append(skipped, s)
			continue
		}
		valid = append(valid, s)
	}
	return
}

func decodeDateNode(n yaml.Node) ([]string, error) {
	out, err := decodeStringOrList(n)
	if err != nil {
		return nil, err
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("missing date")
	}
	return out, nil
}

// decodeStringOrList accepts either a scalar string or a sequence of scalars.
// A missing/null node returns an empty slice without error.
func decodeStringOrList(n yaml.Node) ([]string, error) {
	switch n.Kind {
	case 0:
		return nil, nil
	case yaml.ScalarNode:
		if n.Value == "" {
			return nil, nil
		}
		return []string{n.Value}, nil
	case yaml.SequenceNode:
		out := make([]string, 0, len(n.Content))
		for _, c := range n.Content {
			if c.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("list contains non-scalar")
			}
			out = append(out, c.Value)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unexpected node kind %d", n.Kind)
	}
}

func splitArgs(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		out = append(out, strings.TrimSpace(p))
	}
	return out
}

var callRe = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\(([^)]*)\)$`)

// ParseCall splits "name(arg1, arg2)" into its parts.
func ParseCall(expr string) (MethodCall, error) {
	m := callRe.FindStringSubmatch(strings.TrimSpace(expr))
	if m == nil {
		return MethodCall{}, fmt.Errorf("invalid call expression %q", expr)
	}
	mc := MethodCall{Name: m[1]}
	if strings.TrimSpace(m[2]) != "" {
		for _, a := range strings.Split(m[2], ",") {
			mc.Args = append(mc.Args, strings.TrimSpace(a))
		}
	}
	return mc, nil
}
