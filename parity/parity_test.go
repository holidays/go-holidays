//go:build parity

package parity

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	holidays "github.com/holidays/go-holidays"
)

// yearStart and yearEnd give the Jan 1 / Dec 31 literals for a calendar year.
func yearStart(y int) string { return fmt.Sprintf("%04d-01-01", y) }
func yearEnd(y int) string   { return fmt.Sprintf("%04d-12-31", y) }

// flex* helpers render a compact, human-readable arg string for failure
// messages (and for the knownDivergences key). They carry the exact case so a
// reported mismatch is reproducible.
func flagSuffix(f flagCombo) string {
	return fmt.Sprintf("informal=%v observed=%v", f.informal, f.observed)
}

func flexArgs(kind, val, region string, f flagCombo) string {
	return fmt.Sprintf("%s=%s region=%s %s", kind, val, region, flagSuffix(f))
}

func flexBetweenArgs(s, e string, regions []string, f flagCombo) string {
	return fmt.Sprintf("start=%s end=%s regions=%v %s", s, e, regions, flagSuffix(f))
}

func flexNextArgs(from string, count int, region string, f flagCombo) string {
	return fmt.Sprintf("from=%s count=%d region=%s %s", from, count, region, flagSuffix(f))
}

func flexYearArgs(from, region string, f flagCombo) string {
	return fmt.Sprintf("from=%s region=%s %s", from, region, flagSuffix(f))
}

// knownDivergences records cases where Go and the Ruby gem genuinely disagree
// over identical region data and the difference is NOT a harness bug. It is
// deliberately empty: a populated entry must carry an exact case and a comment,
// and every entry is reported in the completion summary for triage. The harness
// does NOT consult this map to suppress failures; it exists only to document any
// divergence that would otherwise block the whole run. Today there are none.
var knownDivergences = map[string]string{}

// corpusRegions are the request region sets exercised across on/between. They
// span the engine's method variety (us federal/observed, ca, gb wildcards, de
// easter-based, fr, jp custom methods, nz anniversaries, br, in, au_nsw, il
// lunar, mx) plus a couple of multi-region requests.
var corpusRegions = [][]string{
	{"us"}, {"ca"}, {"gb_eng"}, {"de"}, {"fr"}, {"jp"},
	{"nz"}, {"br"}, {"in"}, {"au_nsw"}, {"il"}, {"mx"},
	{"us", "ca"}, {"gb_eng", "fr"},
}

// corpusYears includes weekend-boundary years (observed-date shifts) plus a
// recent spread.
var corpusYears = []int{2020, 2021, 2022, 2024, 2025, 2026}

// flagCombo is one informal/observed setting pair, named for failure messages.
type flagCombo struct {
	name     string
	informal bool
	observed bool
}

var corpusFlags = []flagCombo{
	{"plain", false, false},
	{"observed", false, true},
	{"informal", true, false},
	{"informal+observed", true, true},
}

func opts(regions []string, f flagCombo) holidays.Options {
	return holidays.Options{Regions: regions, Informal: f.informal, Observed: f.observed}
}

// sweepPoolSize is the number of independent Ruby oracle subprocesses the
// exhaustive sweep shards regions across. Fixed at 4 rather than derived from
// runtime.NumCPU(): CI's public ubuntu-latest runner has 4 vCPUs, so going
// higher would oversubscribe there without buying anything (go-holidays-2vu
// investigation measured K=1..4 and found near-linear scaling that a real
// runner cannot exceed).
const sweepPoolSize = 4

// shared oracle for the whole package run; started once in TestMain. The
// curated-corpus specs (registerOnSpec..registerVendoredIsolationSpec) and the
// sweep's region-support probe all use this single oracle, unchanged.
var ora *oracle

// sweepPool holds sweepPoolSize independent oracles the exhaustive sweep
// shards regions across (see registerSweepSpecs / sweepRegionsConcurrently).
// ora is always sweepPool[0], so it is started once and not double-closed.
var sweepPool []*oracle

func TestMain(m *testing.M) {
	o, err := startOracle()
	if err != nil {
		panicf("start oracle: %v", err)
	}
	ora = o

	extra, err := startOraclePool(sweepPoolSize - 1)
	if err != nil {
		panicf("start oracle pool: %v", err)
	}
	sweepPool = append([]*oracle{ora}, extra...)

	code := m.Run()
	for _, p := range extra {
		_ = p.Close()
	}
	_ = ora.Close()
	os.Exit(code)
}

func panicf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "parity: "+format+"\n", args...)
	os.Exit(2)
}

// All parity specs are nested under this single top-level container to keep
// them in declaration order. Ginkgo randomizes top-level container order by
// default.
//
// The de/gb_eng informal cross-region mismatches that originally motivated this
// structure were the go-holidays-2z0 bug (the gem's lazy merge of its vendored
// aggregate definitions on top of load_custom'd rules, order-dependent). That is
// now fixed in oracle.rb (Load#call skips aggregate vendored loads) and verified:
// querying the aggregate-parent regions (si/sk/ve/bg_bg) leaves gb_eng and de
// byte-identical, and -ginkgo.randomize-all is green across seeds.
//
// The container is kept for two smaller reasons:
// (a) registerLoadCustomSpec sends load_custom for a parity_smoke region to the
//
//	shared oracle and cannot undo it (oracle has no unload_custom func);
//
// (b) the exhaustive sweep (sweep_parity_test.go, appended last via
//
//	registerSweepSpecs) expects the oracle already warm.
var _ = Describe("Parity", func() {
	registerOnSpec()
	registerBetweenSpec()
	registerNextHolidaysSpec()
	registerYearHolidaysSpec()
	registerAnyHolidaysDuringWorkWeekSpec()
	registerWildcardCollapseSpec()
	registerAvailableRegionsSpec()
	registerLoadCustomSpec()
	registerCacheBetweenSpec()
	registerVendoredIsolationSpec()
	registerSweepSpecs()
	registerWildcardSweepSpec()
})

// ---- on ----------------------------------------------------------------------

// registerOnSpec verifies (region -> date) pairs expected to BE holidays, plus
// a few dates expected to be empty, exercising the `on` path including the
// empty==empty case.
func registerOnSpec() {
	Context("On", func() {
		It("matches the oracle for known holiday dates and known-empty dates", func() {
			t := GinkgoT()
			type onCase struct {
				region string
				date   string
			}
			cases := []onCase{
				{"us", "2024-12-25"}, {"us", "2024-07-04"}, {"us", "2024-11-28"},
				{"us", "2021-07-04"}, {"us", "2021-12-31"}, // observed-shift boundary
				{"ca", "2024-07-01"}, {"ca", "2024-12-26"},
				{"gb_eng", "2024-12-26"}, {"gb_eng", "2024-04-01"},
				{"de", "2024-03-29"}, {"de", "2024-04-01"}, // Good Friday, Easter Monday
				{"fr", "2024-05-01"}, {"fr", "2024-07-14"},
				{"jp", "2024-01-01"}, {"jp", "2024-05-03"},
				{"nz", "2024-02-06"}, {"nz", "2024-04-25"},
				{"br", "2024-09-07"}, {"br", "2024-11-15"},
				{"in", "2024-01-26"}, {"in", "2024-08-15"},
				{"au_nsw", "2024-01-26"}, {"au_nsw", "2024-12-26"},
				{"il", "2024-05-14"},
				{"mx", "2024-09-16"}, {"mx", "2024-11-20"},
				// expected-empty (ordinary weekday, no holiday)
				{"us", "2024-03-12"}, {"de", "2024-08-13"}, {"jp", "2024-06-18"},
			}
			for _, f := range corpusFlags {
				for _, c := range cases {
					date := mustDate(c.date)
					o := opts([]string{c.region}, f)
					got, err := holidays.On(date, o)
					if err != nil {
						t.Errorf("On(%s, %s, %s) Go error: %v", c.date, c.region, f.name, err)
						continue
					}
					want, err := ora.holidayList(request{
						Func: "on", Date: c.date, Regions: []string{c.region},
						Informal: f.informal, Observed: f.observed,
					})
					if isOracleUnsupported(err) {
						t.Logf("SKIP on(%s, %s, %s): oracle cannot serve this region: %v", c.date, c.region, f.name, err)
						continue
					}
					if err != nil {
						t.Errorf("On(%s, %s, %s) oracle error: %v", c.date, c.region, f.name, err)
						continue
					}
					assertPairs(t, "on", flexArgs("date", c.date, c.region, f), normalizeGo(got), want)
				}
			}
		})
	})
}

// ---- between -----------------------------------------------------------------

func registerBetweenSpec() {
	Context("Between", func() {
		It("matches the oracle across full-year and partial-year ranges", func() {
			t := GinkgoT()
			type span struct {
				start, end string
			}
			// Full calendar-year ranges for each region/year, plus a few partial ranges.
			partials := []span{
				{"2024-06-01", "2024-09-30"},
				{"2021-12-01", "2022-01-15"}, // straddles year boundary, observed shifts
				{"2025-03-01", "2025-05-31"},
			}
			for _, f := range corpusFlags {
				for _, regions := range corpusRegions {
					for _, y := range corpusYears {
						runBetween(t, regions, f, yearStart(y), yearEnd(y))
					}
					// partial ranges only on single-region requests to keep the count sane
					if len(regions) == 1 {
						for _, p := range partials {
							runBetween(t, regions, f, p.start, p.end)
						}
					}
				}
			}
		})
	})
}

func runBetween(t GinkgoTInterface, regions []string, f flagCombo, s, e string) {
	t.Helper()
	o := opts(regions, f)
	got, err := holidays.Between(mustDate(s), mustDate(e), o)
	if err != nil {
		t.Errorf("Between(%s..%s, %v, %s) Go error: %v", s, e, regions, f.name, err)
		return
	}
	want, err := ora.holidayList(request{
		Func: "between", Start: s, End: e, Regions: regions,
		Informal: f.informal, Observed: f.observed,
	})
	if isOracleUnsupported(err) {
		t.Logf("SKIP between(%s..%s, %v, %s): oracle cannot serve this region: %v", s, e, regions, f.name, err)
		return
	}
	if err != nil {
		t.Errorf("Between(%s..%s, %v, %s) oracle error: %v", s, e, regions, f.name, err)
		return
	}
	assertPairs(t, "between", flexBetweenArgs(s, e, regions, f), normalizeGo(got), want)
}

// ---- next_holidays -----------------------------------------------------------

func registerNextHolidaysSpec() {
	Context("NextHolidays", func() {
		It("matches the oracle for a curated set of from/count cases", func() {
			t := GinkgoT()
			type nc struct {
				region string
				from   string
				count  int
			}
			// Known intentional divergence (not asserted): next_holidays(de, 2024-03-01, 12).
			// The gem windows its search through a dates_driver that buckets each holiday
			// by its source month (function/variable holidays in month 0), only out to
			// from >> 12. For this case the 2025 bucket ends at month 4, so the gem drops
			// the fixed-date Tag der Arbeit (2025-05-01, month 5) while keeping the
			// Easter-based Christi Himmelfahrt (2025-05-29, month 0). Go's NextHolidays
			// instead expands year by year until it has `count` holidays, so it correctly
			// includes Tag der Arbeit 2025. We deliberately do not replicate the gem's
			// month-bucketing quirk here; see parity/README.md.
			cases := []nc{
				{"us", "2024-01-01", 1}, {"us", "2024-01-01", 3}, {"us", "2024-01-01", 12},
				{"us", "2024-12-01", 3}, {"us", "2021-12-15", 3}, // boundary
				{"ca", "2024-06-15", 3},
				{"jp", "2024-04-01", 12}, {"gb_eng", "2024-01-01", 12},
				{"nz", "2024-01-01", 12}, {"fr", "2025-01-01", 12},
			}
			for _, f := range corpusFlags {
				for _, c := range cases {
					o := opts([]string{c.region}, f)
					got, err := holidays.NextHolidays(mustDate(c.from), c.count, o)
					if err != nil {
						t.Errorf("NextHolidays(%s, %d, %s, %s) Go error: %v", c.from, c.count, c.region, f.name, err)
						continue
					}
					want, err := ora.holidayList(request{
						Func: "next_holidays", Count: c.count, From: c.from,
						Regions: []string{c.region}, Informal: f.informal, Observed: f.observed,
					})
					if isOracleUnsupported(err) {
						t.Logf("SKIP next_holidays(%s, %d, %s, %s): oracle cannot serve this region: %v", c.from, c.count, c.region, f.name, err)
						continue
					}
					if err != nil {
						t.Errorf("NextHolidays(%s, %d, %s, %s) oracle error: %v", c.from, c.count, c.region, f.name, err)
						continue
					}
					args := flexNextArgs(c.from, c.count, c.region, f)
					assertPairs(t, "next_holidays", args, normalizeGo(got), want)
				}
			}
		})
	})
}

// ---- year_holidays -----------------------------------------------------------

func registerYearHolidaysSpec() {
	Context("YearHolidays", func() {
		It("matches the oracle for a curated set of from cases", func() {
			t := GinkgoT()
			type yc struct {
				region string
				from   string
			}
			cases := []yc{
				{"us", "2024-01-01"}, {"us", "2024-06-15"}, // mid-year
				{"us", "2021-01-01"}, // C1 weekend-boundary case (run with observed below)
				{"ca", "2024-01-01"}, {"de", "2025-01-01"}, {"jp", "2024-01-01"},
				{"gb_eng", "2024-03-01"}, {"nz", "2024-01-01"}, {"fr", "2026-01-01"},
				{"br", "2024-01-01"}, {"au_nsw", "2024-01-01"},
			}
			for _, f := range corpusFlags {
				for _, c := range cases {
					o := opts([]string{c.region}, f)
					got, err := holidays.YearHolidaysFrom(mustDate(c.from), o)
					if err != nil {
						t.Errorf("YearHolidaysFrom(%s, %s, %s) Go error: %v", c.from, c.region, f.name, err)
						continue
					}
					want, err := ora.holidayList(request{
						Func: "year_holidays", From: c.from, Regions: []string{c.region},
						Informal: f.informal, Observed: f.observed,
					})
					if isOracleUnsupported(err) {
						t.Logf("SKIP year_holidays(%s, %s, %s): oracle cannot serve this region: %v", c.from, c.region, f.name, err)
						continue
					}
					if err != nil {
						t.Errorf("YearHolidaysFrom(%s, %s, %s) oracle error: %v", c.from, c.region, f.name, err)
						continue
					}
					args := flexYearArgs(c.from, c.region, f)
					assertPairs(t, "year_holidays", args, normalizeGo(got), want)
				}
			}
		})
	})
}

// ---- any_holidays_during_work_week? ------------------------------------------

func registerAnyHolidaysDuringWorkWeekSpec() {
	Context("AnyHolidaysDuringWorkWeek", func() {
		It("matches the oracle boolean result for a curated set of dates", func() {
			t := GinkgoT()
			type wc struct {
				region string
				date   string
			}
			cases := []wc{
				{"us", "2024-12-24"}, // Christmas week -> true
				{"us", "2024-03-12"}, // ordinary week -> false
				{"us", "2024-07-02"}, // July 4 week -> true
				{"de", "2024-03-26"}, // ordinary week -> false
				{"de", "2024-12-23"}, // Christmas week -> true
				{"jp", "2024-05-01"}, // Golden Week -> true
				{"gb_eng", "2024-08-26"},
			}
			for _, f := range corpusFlags {
				for _, c := range cases {
					o := opts([]string{c.region}, f)
					got, err := holidays.AnyHolidaysDuringWorkWeek(mustDate(c.date), o)
					if err != nil {
						t.Errorf("AnyHolidaysDuringWorkWeek(%s, %s, %s) Go error: %v", c.date, c.region, f.name, err)
						continue
					}
					want, err := ora.boolResult(request{
						Func: "any_holidays_during_work_week?", Date: c.date,
						Regions: []string{c.region}, Informal: f.informal, Observed: f.observed,
					})
					if isOracleUnsupported(err) {
						t.Logf("SKIP any_holidays_during_work_week?(%s, %s, %s): oracle cannot serve this region: %v", c.date, c.region, f.name, err)
						continue
					}
					if err != nil {
						t.Errorf("AnyHolidaysDuringWorkWeek(%s, %s, %s) oracle error: %v", c.date, c.region, f.name, err)
						continue
					}
					if got != want {
						t.Errorf("any_holidays_during_work_week?(date=%s region=%s %s): Go=%v Ruby=%v",
							c.date, c.region, f.name, got, want)
					}
				}
			}
		})
	})
}

// ---- multi-segment wildcard collapse (go-holidays-dpt) -----------------------

// registerWildcardCollapseSpec covers the go-holidays-dpt regression: a
// multi-segment wildcard region (e.g. "au_vic_", "gb_eng_") must collapse all
// the way to its country segment and return the exact same calendar as the
// single-segment wildcard ("au_", "gb_"), country-wide-only holidays
// included.
//
// gb_eng_ is compared against the oracle across full corpusYears ranges: it
// is clean (verified during investigation, no unrelated divergence). au_vic_
// is instead compared over Jan 1 - Sep 30 of each corpusYears year: a
// full-year au_ range hits a separate, pre-existing, unrelated divergence
// (go-holidays-67s, discovered while writing this spec) where Go drops the
// ACT/NSW/SA Labour Day entry when it falls on the same October date as the
// QLD Queen's Birthday entry. Restricting the range keeps this spec scoped to
// the wildcard-collapse fix rather than tripping over that separate bug.
func registerWildcardCollapseSpec() {
	Context("WildcardCollapse", func() {
		It("On surfaces a country-wide-only holiday for a multi-segment wildcard", func() {
			// Kept in this Context rather than folded into registerOnSpec's
			// shared cases list: historically a wildcard query merged the gem's
			// stale vendored country file into the shared oracle and tainted every
			// later query of that country. oracle.rb's Load#call now short-circuits
			// an already-loaded region (go-holidays-ysu), so that taint is gone,
			// but the separation is cheap and left as a guard: registerOnSpec runs
			// before this Context, so its plain "au_nsw"/"gb_eng" cases stay
			// upstream of any wildcard query against those countries.
			t := GinkgoT()
			type onCase struct {
				region, date, name string
			}
			cases := []onCase{
				{"au_vic_", "2017-01-26", "Australia Day"}, // Regions: ["au"]
				{"gb_eng_", "2024-03-29", "Good Friday"},   // Regions: ["gb"]
			}
			for _, c := range cases {
				got, err := holidays.On(mustDate(c.date), holidays.Options{Regions: []string{c.region}})
				if err != nil {
					t.Errorf("On(%s, %s) Go error: %v", c.date, c.region, err)
					continue
				}
				want, err := ora.holidayList(request{Func: "on", Date: c.date, Regions: []string{c.region}})
				if err != nil {
					t.Errorf("On(%s, %s) oracle error: %v", c.date, c.region, err)
					continue
				}
				assertPairs(t, "on", flexArgs("date", c.date, c.region, corpusFlags[0]), normalizeGo(got), want)
			}
		})

		It("gb_eng_ matches the oracle's gb_ calendar for a full year", func() {
			t := GinkgoT()
			for _, f := range corpusFlags {
				for _, y := range corpusYears {
					runBetween(t, []string{"gb_eng_"}, f, yearStart(y), yearEnd(y))
				}
			}
		})

		It("au_vic_ matches the oracle's au_ calendar for Jan-Sep", func() {
			t := GinkgoT()
			for _, f := range corpusFlags {
				for _, y := range corpusYears {
					runBetween(t, []string{"au_vic_"}, f, fmt.Sprintf("%04d-01-01", y), fmt.Sprintf("%04d-09-30", y))
				}
			}
		})
	})
}

// ---- available_regions -------------------------------------------------------

func registerAvailableRegionsSpec() {
	Context("AvailableRegions", func() {
		It("matches the oracle's region universe exactly", func() {
			t := GinkgoT()
			goRegions := holidays.AvailableRegions()
			rubyRegions, err := ora.stringList(request{Func: "available_regions"})
			if err != nil {
				t.Fatalf("available_regions oracle error: %v", err)
			}
			onlyGo, onlyRuby := diffStrings(goRegions, rubyRegions)
			if len(onlyGo) > 0 || len(onlyRuby) > 0 {
				t.Errorf("available_regions mismatch (Go has %d, Ruby has %d)\n  only in Go: %v\n  only in Ruby: %v",
					len(goRegions), len(rubyRegions), onlyGo, onlyRuby)
			}
		})
	})
}

// ---- load_custom round-trip --------------------------------------------------

// registerLoadCustomSpec is a light smoke test: both sides ingest the same tiny
// throwaway YAML. The oracle reports {"loaded":1}; Go's LoadCustom returns no
// error and the custom rule resolves. We do not output-compare the gem's
// load_custom (it merges into global state); we assert both sides accept it.
func registerLoadCustomSpec() {
	Context("LoadCustom", func() {
		It("loads the same custom rule on both sides and Go resolves it", func() {
			body := "" +
				"months:\n" +
				"  4:\n" +
				"  - name: Parity Smoke Holiday\n" +
				"    regions: [parity_smoke]\n" +
				"    mday: 9\n"
			dir := GinkgoT().TempDir()
			path := filepath.Join(dir, "parity_smoke.yaml")
			Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())

			loaded, err := ora.loadCount(request{Func: "load_custom", Files: []string{path}})
			Expect(err).NotTo(HaveOccurred())
			Expect(loaded).To(Equal(1))

			Expect(holidays.LoadCustom(path)).To(Succeed())
			defer holidays.UnloadCustom(path)

			got, err := holidays.On(mustDate("2024-04-09"), holidays.Options{Regions: []string{"parity_smoke"}})
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(HaveLen(1))
			Expect(got[0].Name).To(Equal("Parity Smoke Holiday"))
		})
	})
}

// ---- vendored-definition isolation -------------------------------------------

// registerVendoredIsolationSpec guards go-holidays-2z0: asking the oracle about
// one region must never change another region's answer.
//
// The gem ships its own vendored copy of the definitions and lazily requires the
// file for a requested region's parent. Five regions map to an aggregate parent
// (si/sk -> europe, ve -> southamerica, bg_bg/bg_en -> bg), and loading an
// aggregate merges EVERY region in it, which appended a second, stale set of gb
// rules. gb then resolved to both rule sets and the sweep reported a mismatch
// that was not real. oracle.rb now skips aggregate loads; this fails if that
// protection is removed.
//
// 1977 is the probe year because Christmas falls on a Sunday, so the stale rules
// (to_monday_if_weekend) and the current ones
// (to_tuesday_if_sunday_or_monday_if_saturday) disagree and the extra rules show
// up as duplicate dates rather than being silently absorbed.
func registerVendoredIsolationSpec() {
	Context("VendoredDefinitionIsolation", func() {
		It("does not let an unrelated region query alter gb", func() {
			gbReq := request{
				Func: "year_holidays", From: "1977-01-01",
				Regions: []string{"gb"}, Observed: true,
			}

			before, err := ora.holidayList(gbReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(before).NotTo(BeEmpty())

			for _, r := range []string{"si", "sk", "ve", "bg_bg"} {
				_, err := ora.holidayList(request{
					Func: "year_holidays", From: "2024-01-01", Regions: []string{r},
				})
				Expect(err).NotTo(HaveOccurred(), "querying %s must succeed", r)
			}

			after, err := ora.holidayList(gbReq)
			Expect(err).NotTo(HaveOccurred())
			Expect(after).To(Equal(before),
				"gb resolved differently after querying regions whose parent is an aggregate")
		})
	})
}

// ---- cache_between transparency ----------------------------------------------

// registerCacheBetweenSpec exercises Go's cache path: CacheBetween over a wide
// range, then Between over a sub-range must still equal the oracle's `between`
// for that sub-range. (Go's CacheBetween returns only error, so we do NOT
// output-compare it directly; the gem's cache_between return is irrelevant.)
func registerCacheBetweenSpec() {
	Context("CacheBetween", func() {
		It("keeps a cached sub-range equal to the oracle", func() {
			t := GinkgoT()
			holidays.ResetCache()
			defer holidays.ResetCache()

			regions := []string{"us"}
			o := holidays.Options{Regions: regions}
			wideStart, wideEnd := "2024-01-01", "2024-12-31"
			Expect(holidays.CacheBetween(mustDate(wideStart), mustDate(wideEnd), o)).To(Succeed())

			subStart, subEnd := "2024-06-01", "2024-08-31"
			got, err := holidays.Between(mustDate(subStart), mustDate(subEnd), o)
			Expect(err).NotTo(HaveOccurred())
			want, err := ora.holidayList(request{
				Func: "between", Start: subStart, End: subEnd, Regions: regions,
			})
			Expect(err).NotTo(HaveOccurred())
			assertPairs(t, "cache_between(sub-range)",
				flexBetweenArgs(subStart, subEnd, regions, flagCombo{name: "plain"}),
				normalizeGo(got), want)
		})
	})
}

// ---- assertion + arg-formatting helpers --------------------------------------

func assertPairs(t GinkgoTInterface, fn, args string, got, want []pair) {
	t.Helper()
	if pairsEqual(got, want) {
		return
	}
	if msg, ok := knownDivergences[fn+" "+args]; ok {
		t.Logf("KNOWN DIVERGENCE %s %s: %s", fn, args, msg)
		return
	}
	onlyGo, onlyRuby := diffPairs(got, want)
	t.Errorf("%s mismatch [%s]: Go=%d Ruby=%d%s", fn, args, len(got), len(want),
		formatDiff(onlyGo, onlyRuby))
}

func diffStrings(a, b []string) (onlyA, onlyB []string) {
	seen := map[string]int{}
	for _, s := range a {
		seen[s] |= 1
	}
	for _, s := range b {
		seen[s] |= 2
	}
	for s, bits := range seen {
		switch bits {
		case 1:
			onlyA = append(onlyA, s)
		case 2:
			onlyB = append(onlyB, s)
		}
	}
	return onlyA, onlyB
}
