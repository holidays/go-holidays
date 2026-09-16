package engine

import (
	"fmt"
	"time"

	"github.com/holidays/go-holidays/internal/definition"
)

// Resolved is the internal result of resolving a single rule for a single year.
type Resolved struct {
	Date     time.Time
	Name     string
	Regions  []string
	Informal bool
}

// ResolveOptions controls filtering during resolution.
type ResolveOptions struct {
	Regions  []string
	Informal bool
	Observed bool
}

// ResolveYear walks the registered rules and returns holidays for the given year.
// Rules that are identical except for their regions are de-duplicated, keeping
// the first occurrence, so a multi-region request whose regions each define the
// same holiday (for example informal Easter Sunday in both us and ca) yields it
// once. Two definitions merge when they match on name, wday, mday, week,
// function, function_modifier, type, observed, and year_ranges; definitions that
// differ on any of those fields (for example us Good Friday tagged informal vs
// the untagged us-states/ca Good Friday) stay distinct.
func ResolveYear(year int, opts ResolveOptions) ([]Resolved, error) {
	var (
		rules = rulesFor(opts.Regions)
		out   = make([]Resolved, 0, len(rules))
		seen  = make(map[string]struct{}, len(rules))
	)
	for _, rule := range rules {
		if !rule.AppliesIn(year) {
			continue
		}
		if rule.Type == definition.Informal && !opts.Informal {
			continue
		}
		if !ruleMatchesRequested(rule, opts.Regions) {
			continue
		}
		date, err := computeDate(rule, year)
		if err != nil {
			return nil, fmt.Errorf("rule %q: %w", rule.Name, err)
		}
		if date.IsZero() {
			continue
		}
		if opts.Observed && rule.HasObserved() {
			date, err = applyObserved(rule.Observed, date, requestedRegion(opts.Regions))
			if err != nil {
				return nil, fmt.Errorf("rule %q observed=%s: %w", rule.Name, rule.Observed, err)
			}
		}
		key := ruleSignature(rule)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, Resolved{Date: date, Name: rule.Name, Regions: rule.Regions, Informal: rule.Type == definition.Informal})
	}
	return out, nil
}

// ruleSignature renders the fields that decide whether two definitions are
// the same holiday (and merges their regions): month, name, wday, mday, week,
// function, function_modifier, type, observed, and year_ranges. Month is part
// of the identity because definitions are bucketed per month, so two rules in
// different months never collide even when every other field matches (for
// example the WA "Labour Day" first Monday of March and the ACT/NSW/SA
// "Labour Day" first Monday of October). Regions are deliberately excluded, so
// rules that differ only by region collapse to one Resolved.
func ruleSignature(r definition.HolidayRule) string {
	return fmt.Sprintf("%d|%s|%d|%d|%d|%s|%d|%d|%s|%v",
		r.Month, r.Name, r.Wday, r.Mday, r.Week, r.Function, r.FunctionModifier,
		r.Type, r.Observed, r.YearRanges)
}

func computeDate(rule definition.HolidayRule, year int) (time.Time, error) {
	var base time.Time
	switch {
	case rule.HasMday():
		base = time.Date(year, time.Month(rule.Month), rule.Mday, 0, 0, 0, 0, time.UTC)
	case rule.HasWday():
		base = nthOrLastWeekday(year, time.Month(rule.Month), rule.Week, time.Weekday(rule.Wday))
		if base.IsZero() {
			return time.Time{}, nil
		}
	}
	if rule.HasFunction() {
		fn, ok := LookupMethod(rule.Function)
		if !ok {
			return time.Time{}, fmt.Errorf("unregistered method %q", rule.Function)
		}
		args := MethodArgs{Year: year, Month: rule.Month, Day: rule.Mday, Date: base}
		if len(rule.Regions) > 0 {
			args.Region = rule.Regions[0]
		}
		d, err := fn(args)
		if err != nil {
			return time.Time{}, err
		}
		base = d
	}
	if base.IsZero() {
		if !rule.HasMday() && !rule.HasWday() && !rule.HasFunction() {
			return time.Time{}, fmt.Errorf("rule has no mday, wday/week, or function")
		}
		return time.Time{}, nil
	}
	if rule.FunctionModifier != 0 {
		base = base.AddDate(0, 0, rule.FunctionModifier)
	}
	if rule.HasFunction() && !base.IsZero() {
		// For a function holiday, keep the computed month/day but force the
		// resolution year. This is a no-op for mid-year results (they already
		// fall in `year`); it pulls a lunar month-12 eve, whose solar date lands
		// in the next gregorian year, back into `year`, so a region like kr emits
		// both Seollal holiday days (eve + day-after).
		base = time.Date(year, base.Month(), base.Day(), 0, 0, 0, 0, time.UTC)
	}
	return base, nil
}

// requestedRegion picks the region an observed method sees. The observed
// method's input leaves the holiday's own regions out, so the region is the
// first region the caller ASKED for, not the first the rule declares. That
// distinction is what lets one shared definition observe differently in a
// sub-region: us Juneteenth is defined once for [us], and only a us_ut request
// gets Utah's substitution rule. An all-regions request has no queried region.
func requestedRegion(regions []string) string {
	if len(regions) == 0 {
		return ""
	}
	return regions[0]
}

func applyObserved(methodName string, date time.Time, region string) (time.Time, error) {
	fn, ok := LookupMethod(methodName)
	if !ok {
		return time.Time{}, fmt.Errorf("unregistered observed method %q", methodName)
	}
	return fn(MethodArgs{
		Date:   date,
		Year:   date.Year(),
		Month:  int(date.Month()),
		Day:    date.Day(),
		Region: region,
	})
}

func nthOrLastWeekday(year int, month time.Month, week int, wday time.Weekday) time.Time {
	if week == -1 {
		firstNext := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)
		last := firstNext.AddDate(0, 0, -1)
		offset := (int(last.Weekday()) - int(wday) + 7) % 7
		return last.AddDate(0, 0, -offset)
	}
	first := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	offset := (int(wday) - int(first.Weekday()) + 7) % 7
	day := 1 + offset + (week-1)*7
	candidate := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	if candidate.Month() != month {
		return time.Time{}
	}
	return candidate
}
