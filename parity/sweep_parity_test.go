//go:build parity

package parity

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	. "github.com/onsi/ginkgo/v2"

	holidays "github.com/holidays/go-holidays"
)

// sweepHeartbeatInterval is how often the exhaustive sweep prints a progress
// line. The per-region oracle call is multi-second Ruby compute and Ginkgo
// buffers GinkgoWriter until the spec ends, so without this the run is silent
// for minutes. The line goes straight to stdout to bypass that buffering.
const sweepHeartbeatInterval = 15 * time.Second

// startSweepHeartbeat spawns a goroutine that prints "[parity sweep] N/total,
// Ns elapsed" every sweepHeartbeatInterval until the returned stop func is
// called. done is read atomically by the goroutine; the caller bumps it.
func startSweepHeartbeat(label string, done *int64, total int) (stop func()) {
	start := time.Now()
	ticker := time.NewTicker(sweepHeartbeatInterval)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-quit:
				return
			case <-ticker.C:
				fmt.Fprintf(os.Stdout, "[parity sweep] %s: %d/%d, %s elapsed\n",
					label, atomic.LoadInt64(done), total, time.Since(start).Round(time.Second))
			}
		}
	}()
	return func() {
		ticker.Stop()
		close(quit)
	}
}

// sweepYearStart and sweepYearEnd bound the exhaustive year_holidays sweep. The
// range spans weekend-boundary years (observed-date shifts), historical years,
// and well into the future so year_range and leap-year paths are all exercised.
const (
	sweepYearStart = 1970
	sweepYearEnd   = 2050
)

// maxReportedMismatches caps the per-run failure spew: the first N mismatches
// are reported in full; the rest are only counted, so a systemic break does not
// bury the summary under tens of thousands of lines.
const maxReportedMismatches = 50

// registerSweepSpecs is the broad equivalence proof. For every region the
// oracle can serve (all of them today; a shim in oracle.rb keeps jp serviceable
// despite the gem's ported jp_next_weekday needing a Ruby-coded module), it
// compares the full year_holidays list from Go's YearHolidaysFrom against the
// oracle for every year in [sweepYearStart, sweepYearEnd] under all four flag
// combinations.
//
// It first enumerates which of AvailableRegions() the oracle can serve and
// logs an explicit coverage tally, so skipped regions are accounted for rather
// than silently passing. Unlike the curated corpus, this exercises every region
// and a wide year span, turning "we checked a few cases" into "every holiday
// each engine produces, for every serviceable region and year, matches".
//
// Called from the single top-level Describe in parity_test.go (see the
// comment there) so this spec keeps running last, after the shared oracle has
// already been exercised by the curated-corpus specs, matching the original
// go test file order.
func registerSweepSpecs() {
	Context("ExhaustiveYearSweep", func() {
		It("matches the oracle for every serviceable region and year in range", func() {
			t := GinkgoT()
			regions := availableRegionsSorted(t)

			fmt.Fprintf(os.Stdout, "[parity sweep] probing %d regions for oracle support\n", len(regions))
			var probed int64
			stopProbe := startSweepHeartbeat("probe", &probed, len(regions))
			serviceable, unsupported := partitionByOracleSupport(t, regions, &probed)
			stopProbe()
			// Echoed to stdout as well as the buffered spec log: it frames the
			// per-region heartbeats that follow with the count they run against.
			fmt.Fprintf(os.Stdout, "[parity sweep] coverage: %d regions, %d serviceable, %d oracle-unsupported\n",
				len(regions), len(serviceable), len(unsupported))
			t.Logf("coverage: %d regions total, %d serviceable, %d oracle-unsupported",
				len(regions), len(serviceable), len(unsupported))
			if len(unsupported) > 0 {
				t.Logf("oracle-unsupported regions (skipped, e.g. gem needs a Ruby-coded module): %v", unsupported)
			}

			years := sweepYearEnd - sweepYearStart + 1
			var c sweepCounters
			// Shard serviceable regions across sweepPool (sweepPoolSize independent
			// Ruby oracle subprocesses; ora is sweepPool[0]) and sweep each shard
			// concurrently. Older gems (9.1.0) accumulated cross-region global-state
			// pollution across many load_custom queries, which once forced a fresh
			// per-region oracle (spawn ruby + reload all 80 YAML, ~287x) and made
			// this sweep take minutes. Gem 9.1.2 (#344, region load-order) and
			// 10.0.0 (#352, function_modifier merge) fixed that, so each pool member
			// is warmed once and reused for its whole shard. sweepCounters.mu makes
			// the shared tally safe for the concurrent writers below.
			var swept int64
			stopSweep := startSweepHeartbeat("regions", &swept, len(serviceable))
			sweepRegionsConcurrently(t, sweepPool, serviceable, &c, &swept)
			stopSweep()

			fmt.Fprintf(os.Stdout, "[parity sweep] done: %d comparisons, %d mismatches, %d mutual lunar-boundary confirmations\n",
				c.comparisons, c.mismatches, c.lunarBoundary)
			t.Logf("exhaustive year sweep: %d comparisons across %d serviceable regions x %d years x %d flag combos; %d mismatches, %d mutual lunar-boundary confirmations",
				c.comparisons, len(serviceable), years, len(corpusFlags), c.mismatches, c.lunarBoundary)
			for _, msg := range c.failMessages {
				t.Log(msg)
			}
			if c.failCount > 0 {
				t.Errorf("year sweep: %d failures total (%d reported above); see messages and summary tally",
					c.failCount, len(c.failMessages))
			}
		})
	})
}

// sweepWildcardRegions is a curated representative sample for the
// wildcard-collapse sweep (go-holidays-dpt): one subregion per country that
// has subregions, turned into its wildcard form (e.g. "de_bb_"). Every one of
// AvailableRegions()'s 214 underscore-containing regions is a valid Go-side
// target (the wildcard collapse is per-country, not per-subregion, so any
// subregion under a given country exercises the same code path as any
// other), but sweeping all 214 would roughly double this sweep's oracle
// round-trips for no extra coverage once one region per country has proven
// the collapse. Four country prefixes are deliberately left out:
//   - be_*, bg_*, mt_*, rs_*: these countries have no bare-ISO region. Only
//     language or script sub-variants exist (be_fr, bg_bg, mt_en, rs_cyrl),
//     with no generated parent, so the gem's wildcard-base derivation
//     (wildcard_region.split('_').first) yields be/bg/mt/rs, which is not a
//     valid region, and the gem raises Holidays::InvalidRegion. This is a
//     permanent gem limitation, not an oracle setup gap. Go does not error:
//     it prefix-matches and returns both sub-variants.
var sweepWildcardRegions = []string{
	"au_vic_", "ca_ab_", "ch_ag_", "de_bb_", "es_an_", "fr_a_", "gb_con_",
	"in_ap_", "it_bl_", "mx_pue_", "nz_ak_", "pt_li_", "us_ak_",
}

// registerWildcardSweepSpec extends the exhaustive sweep to the wildcard
// region form (go-holidays-dpt): for each entry in sweepWildcardRegions, it
// compares Go's YearHolidaysFrom against the oracle across the same full
// [sweepYearStart, sweepYearEnd] year span and all four flag combinations
// that registerSweepSpecs uses for bare region codes, reusing the same
// sweepRegion/compareYear machinery (including its set-based, deduped
// comparison). oracle.rb passes any region string through generically (raw
// string -> to_sym, no allowlist), wildcard suffix included, so this is pure
// Go-side test coverage.
func registerWildcardSweepSpec() {
	Context("ExhaustiveWildcardSweep", func() {
		It("matches the oracle for a representative sample of wildcard-collapsed regions", func() {
			t := GinkgoT()
			years := sweepYearEnd - sweepYearStart + 1
			var c sweepCounters
			var swept int64
			stopSweep := startSweepHeartbeat("wildcard regions", &swept, len(sweepWildcardRegions))
			sweepRegionsConcurrently(t, sweepPool, sweepWildcardRegions, &c, &swept)
			stopSweep()

			fmt.Fprintf(os.Stdout, "[parity sweep] wildcard done: %d comparisons, %d mismatches, %d mutual lunar-boundary confirmations\n",
				c.comparisons, c.mismatches, c.lunarBoundary)
			t.Logf("exhaustive wildcard sweep: %d comparisons across %d regions x %d years x %d flag combos; %d mismatches, %d mutual lunar-boundary confirmations",
				c.comparisons, len(sweepWildcardRegions), years, len(corpusFlags), c.mismatches, c.lunarBoundary)
			for _, msg := range c.failMessages {
				t.Log(msg)
			}
			if c.failCount > 0 {
				t.Errorf("wildcard sweep: %d failures total (%d reported above); see messages and summary tally",
					c.failCount, len(c.failMessages))
			}
		})
	})
}

// sweepCounters accumulates the sweep tally across all regions. Once
// sweepRegionsConcurrently runs region shards on separate goroutines (one per
// pool oracle), every field here is written from multiple goroutines, so mu
// guards every mutation (see recordFailure and the increments in
// compareYear).
type sweepCounters struct {
	mu                                     sync.Mutex
	comparisons, mismatches, lunarBoundary int

	// failCount is uncapped; failMessages holds up to maxReportedMismatches.
	failCount    int
	failMessages []string
}

// recordFailure accumulates one compareYear failure without aborting the
// sweep. Only the first maxReportedMismatches messages are retained; failCount
// stays uncapped so the final summary reports the true total.
func (c *sweepCounters) recordFailure(format string, args ...any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failCount++
	if len(c.failMessages) < maxReportedMismatches {
		c.failMessages = append(c.failMessages, fmt.Sprintf(format, args...))
	}
}

// addComparison records one completed comparison (matching or not).
func (c *sweepCounters) addComparison() {
	c.mu.Lock()
	c.comparisons++
	c.mu.Unlock()
}

// addMismatch records one comparison that also mismatched.
func (c *sweepCounters) addMismatch() {
	c.mu.Lock()
	c.mismatches++
	c.mu.Unlock()
}

// addLunarBoundary records one mutual lunar-table boundary confirmation.
func (c *sweepCounters) addLunarBoundary() {
	c.mu.Lock()
	c.lunarBoundary++
	c.mu.Unlock()
}

// sweepRegionsConcurrently shards regions across pool (round-robin so each
// pool member gets a contiguous, roughly equal-sized run of regions) and
// sweeps each shard on its own goroutine against its own oracle. swept is
// bumped atomically for the heartbeat; c is safe for concurrent use.
func sweepRegionsConcurrently(t GinkgoTInterface, pool []*oracle, regions []string, c *sweepCounters, swept *int64) {
	t.Helper()
	shards := make([][]string, len(pool))
	for i, region := range regions {
		shard := i % len(pool)
		shards[shard] = append(shards[shard], region)
	}

	var wg sync.WaitGroup
	for i, shard := range shards {
		if len(shard) == 0 {
			continue
		}
		wg.Add(1)
		go func(ro *oracle, regions []string) {
			defer wg.Done()
			for _, region := range regions {
				sweepRegion(t, ro, region, c)
				atomic.AddInt64(swept, 1)
			}
		}(pool[i], shard)
	}
	wg.Wait()
}

// sweepRegion compares Go against the shared oracle for one region across the
// full year span and all flag combinations.
func sweepRegion(t GinkgoTInterface, ro *oracle, region string, c *sweepCounters) {
	t.Helper()
	// One batched oracle round-trip for the whole region (year x flag combos),
	// instead of one per (year, flag). Transport-only: compareYear still does the
	// deduped (date, name) set comparison against these cached results.
	combos := make([]flagPair, len(corpusFlags))
	for i, f := range corpusFlags {
		combos[i] = flagPair{Informal: f.informal, Observed: f.observed}
	}
	cache, err := ro.yearHolidaysRange(region, sweepYearStart, sweepYearEnd, combos)
	if err != nil {
		c.recordFailure("year_holidays_range(%s) oracle error: %v", region, err)
		return
	}
	for year := sweepYearStart; year <= sweepYearEnd; year++ {
		from := fmt.Sprintf("%04d-01-01", year)
		for _, f := range corpusFlags {
			compareYear(t, region, year, from, f, cache, c)
		}
	}
}

// compareYear runs one (region, year, flag) comparison and updates the tally,
// reading the oracle's answer from the region's batched result cache.
func compareYear(t GinkgoTInterface, region string, year int, from string, f flagCombo, cache map[sweepKey]sweepEntry, c *sweepCounters) {
	t.Helper()
	entry, ok := cache[sweepKey{year: year, informal: f.informal, observed: f.observed}]
	if !ok {
		c.recordFailure("year sweep [region=%s year=%d %s]: no oracle entry in batched result",
			region, year, f.name)
		return
	}
	hs, err := holidays.YearHolidaysFrom(mustDate(from), opts([]string{region}, f))
	if err != nil {
		// Go's lunar tables stop at 2049, so lunar regions (hk, kr, vn) cannot
		// resolve the primary year 2050. This is a genuine MUTUAL boundary: the
		// gem's tables end at the same year, so Ruby is equally undefined here.
		// Rather than blindly skipping, confirm it: query the oracle and require
		// it to ALSO error. If Ruby instead resolves the year, Go has a real gap.
		// (go-holidays-egm; the year+1 look-ahead at 2049 is handled in Go's
		// YearHolidaysFrom, so only the 2050 primary year reaches here.)
		if isLunarRangeError(err) {
			if entry.oracleErr != "" {
				c.addLunarBoundary()
				return
			}
			c.recordFailure("lunar boundary [region=%s year=%d %s]: Go errored (%v) but Ruby resolved the year",
				region, year, f.name, err)
			return
		}
		c.recordFailure("YearHolidaysFrom(%s, %s, %s) Go error: %v", from, region, f.name, err)
		return
	}
	// Compare Go and oracle output as (date, name) SETS, not multisets. The gem
	// emits benign duplicate rows for some region/flag combos (e.g. br informal
	// years: Ruby yields 13 rows to Go's 12 for the same distinct holidays), an
	// artifact of load_custom merge semantics, not a genuinely missing or extra
	// holiday. This dedupe is empirically required: dropping it surfaces 7290 such
	// spurious mismatches. It is unrelated to the (now removed) per-region oracle.
	got := dedupePairs(normalizeGo(hs))
	if entry.oracleErr != "" {
		// A region serviceable at the probe year should stay serviceable; treat any
		// later oracle error as a real failure to investigate.
		c.recordFailure("year_holidays(%s, %s, %s) oracle error: %v", from, region, f.name, entry.oracleErr)
		return
	}
	want := dedupePairs(entry.pairs)
	c.addComparison()
	if pairsEqual(got, want) {
		return
	}
	onlyGo, onlyRuby := diffPairs(got, want)
	c.addMismatch()
	c.recordFailure("year sweep mismatch [region=%s year=%d %s]: Go=%d Ruby=%d%s",
		region, year, f.name, len(got), len(want), formatDiff(onlyGo, onlyRuby))
}

// isLunarRangeError reports whether a Go error is the lunar-table upper-bound
// limit (LunarToSolar out of range), which lunar regions hit at the primary year
// 2050. The gem's tables end at the same year, so compareYear confirms it as a
// mutual boundary rather than treating it as a Go-only failure.
func isLunarRangeError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "calc.LunarToSolar") &&
		strings.Contains(err.Error(), "out of range")
}

// dedupePairs collapses a sorted pair slice to distinct (date, name) entries, so
// compareYear can compare Go and oracle output as sets; see that call site for
// why the gem's benign duplicate rows make this necessary.
func dedupePairs(ps []pair) []pair {
	if len(ps) < 2 {
		return ps
	}
	out := ps[:1]
	for _, p := range ps[1:] {
		if p != out[len(out)-1] {
			out = append(out, p)
		}
	}
	return out
}

// The sweep carried an allowlist here for (region, holiday) pairs that genuinely
// disagreed between Go and the gem. Its last entry was ph National Heroes Day,
// where Go was intentionally correct (last Monday of August) against a buggy
// ph.yaml. Upstream fixed that in holidays/definitions#345, which shipped in
// v8.0.2 as function: ph_heroes_day(year) with a matching implementation in gem
// 11.2.0, so both engines now agree and the sweep tolerates nothing.

// availableRegionsSorted returns the Go region universe, sorted for stable
// iteration and reporting.
func availableRegionsSorted(t GinkgoTInterface) []string {
	t.Helper()
	regions := append([]string(nil), holidays.AvailableRegions()...)
	sort.Strings(regions)
	return regions
}

// partitionByOracleSupport probes each region once (a single year_holidays
// call) and splits the set into regions the oracle can serve and those it
// cannot (the gem raising "uninitialized constant Holidays::...").
func partitionByOracleSupport(t GinkgoTInterface, regions []string, probed *int64) (serviceable, unsupported []string) {
	t.Helper()
	const probeFrom = "2024-01-01"
	for _, region := range regions {
		_, err := ora.holidayList(request{Func: "year_holidays", From: probeFrom, Regions: []string{region}})
		atomic.AddInt64(probed, 1)
		switch {
		case isOracleUnsupported(err):
			unsupported = append(unsupported, region)
		case err != nil:
			t.Errorf("probe year_holidays(%s, %s) oracle error: %v", probeFrom, region, err)
		default:
			serviceable = append(serviceable, region)
		}
	}
	return serviceable, unsupported
}
