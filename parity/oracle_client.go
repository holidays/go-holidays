//go:build parity

// Package parity drives the Ruby oracle (parity/oracle.rb) as a subprocess and
// compares its output, computed over the same region YAML, against the
// Go holidays port. This file holds the subprocess plumbing and normalization;
// the corpus and assertions live in parity_test.go. Everything is behind the
// `parity` build tag so it never runs under plain `make test`.
package parity

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	holidays "github.com/holidays/go-holidays"
)

// pair is one normalized holiday: a UTC calendar date and a name. Both the Go
// side and the oracle side collapse to a sorted slice of these so the two
// engines can be compared apples to apples (regions and other fields dropped).
type pair struct {
	Date string `json:"date"`
	Name string `json:"name"`
}

// request is one NDJSON request line sent to the oracle's stdin. Fields are
// omitted when empty so each func only carries the arguments it needs; the
// oracle reads exactly one JSON object per line.
type request struct {
	ID         int        `json:"id"`
	Func       string     `json:"func"`
	Date       string     `json:"date,omitempty"`
	Start      string     `json:"start,omitempty"`
	End        string     `json:"end,omitempty"`
	From       string     `json:"from,omitempty"`
	Count      int        `json:"count,omitempty"`
	Regions    []string   `json:"regions,omitempty"`
	Files      []string   `json:"files,omitempty"`
	Informal   bool       `json:"informal,omitempty"`
	Observed   bool       `json:"observed,omitempty"`
	Region     string     `json:"region,omitempty"`
	FromYear   int        `json:"from_year,omitempty"`
	ToYear     int        `json:"to_year,omitempty"`
	FlagCombos []flagPair `json:"flag_combos,omitempty"`
}

// flagPair is one informal/observed setting sent in a year_holidays_range batch.
type flagPair struct {
	Informal bool `json:"informal"`
	Observed bool `json:"observed"`
}

// response is one NDJSON response line. result is left raw and decoded by the
// caller into the shape that func returns (holiday list, bool, region list, or
// the load_custom envelope).
type response struct {
	OK     bool            `json:"ok"`
	ID     int             `json:"id"`
	Func   string          `json:"func"`
	Result json.RawMessage `json:"result"`
	Error  string          `json:"error"`
}

// oracle is a single long-lived Ruby subprocess. It loads all YAML at startup
// (slow), so the harness starts it once and serializes requests over the same
// stdin/stdout pipes for the whole run. nextID makes ids unique per request.
type oracle struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	out    *bufio.Reader
	stderr *strings.Builder
	mu     sync.Mutex
	nextID int
}

// repoRoot resolves the repository root from this source file's directory
// (parity/ lives directly under the root), so the oracle path is stable no
// matter what working directory `go test` runs from.
func repoRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	// `go test` runs in the package directory (parity/); the repo root is its
	// parent. Walk up until we find parity/oracle.rb to be robust either way.
	dir := wd
	for i := 0; i < 6; i++ {
		if _, err := os.Stat(filepath.Join(dir, "parity", "oracle.rb")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("could not locate parity/oracle.rb from %s", wd)
}

// startOracle launches `ruby parity/oracle.rb`, wired for line-delimited JSON,
// and returns a handle the caller drives with call(). The caller must Close it.
func startOracle() (*oracle, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}
	scriptRel := filepath.Join("parity", "oracle.rb")
	cmd := exec.Command("ruby", scriptRel)
	cmd.Dir = root

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start oracle: %w", err)
	}
	return &oracle{
		cmd:    cmd,
		stdin:  stdin,
		out:    bufio.NewReaderSize(stdout, 1<<20),
		stderr: &stderr,
	}, nil
}

// startOraclePool launches n independent `ruby parity/oracle.rb` subprocesses
// for the exhaustive sweep's region shards to run against concurrently. Each
// pool member is a freshly spawned process, NOT forked/cloned from an already
// warmed oracle: measurement during the go-holidays-2vu investigation showed
// forking a warmed oracle is slower than spawning fresh (COW + GC storms), so
// this always starts n new subprocesses. Per-process gem state (Factory,
// ProcResultCache, the JP shim, the aggregate-vendored-load guard) is already
// isolated per oracle.rb instance, so sharding regions across independent
// processes is safe. On partial failure, already-started members are closed
// before returning the error.
func startOraclePool(n int) ([]*oracle, error) {
	pool := make([]*oracle, 0, n)
	for i := 0; i < n; i++ {
		o, err := startOracle()
		if err != nil {
			for _, p := range pool {
				_ = p.Close()
			}
			return nil, fmt.Errorf("start oracle pool member %d/%d: %w", i+1, n, err)
		}
		pool = append(pool, o)
	}
	return pool, nil
}

// call sends one request and reads exactly one response line back, in order.
// It serializes concurrent callers so the NDJSON request/response pairing stays
// 1:1. A non-ok response or an id mismatch is surfaced as an error.
func (o *oracle) call(req request) (*response, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.nextID++
	req.ID = o.nextID

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if _, err := o.stdin.Write(append(body, '\n')); err != nil {
		return nil, fmt.Errorf("write request: %w (stderr: %s)", err, o.stderr.String())
	}

	line, err := o.out.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read response: %w (stderr: %s)", err, o.stderr.String())
	}
	var resp response
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("decode response %q: %w", string(line), err)
	}
	if resp.ID != req.ID {
		return nil, fmt.Errorf("response id %d != request id %d (out of sync)", resp.ID, req.ID)
	}
	if !resp.OK {
		return nil, fmt.Errorf("oracle %s error: %s", req.Func, resp.Error)
	}
	return &resp, nil
}

// isOracleUnsupported reports whether an oracle error reflects a region the
// gem cannot serve from our region YAML (rather than a real mismatch). No
// region hits this today: jp used to, because the gem's ported jp_next_weekday
// reaches for Holidays::JP.holidays_by_month, a module the gem only defines
// when it loads its own precompiled jp file, which load_custom never does; a
// shim in oracle.rb now rebuilds that module from our own loaded rules. The
// check and the per-case guards stay as a harmless safety net for any future
// region whose gem method needs a Ruby-coded module.
func isOracleUnsupported(err error) bool {
	return err != nil && strings.Contains(err.Error(), "uninitialized constant Holidays::")
}

// holidayList runs a request whose result is a holiday list and decodes it into
// a sorted slice of pairs (the oracle already sorts; we sort again defensively).
func (o *oracle) holidayList(req request) ([]pair, error) {
	resp, err := o.call(req)
	if err != nil {
		return nil, err
	}
	var pairs []pair
	if err := json.Unmarshal(resp.Result, &pairs); err != nil {
		return nil, fmt.Errorf("decode holiday list for %s: %w", req.Func, err)
	}
	sortPairs(pairs)
	return pairs, nil
}

// sweepKey identifies one (year, flag-combo) cell of a year_holidays_range batch.
type sweepKey struct {
	year               int
	informal, observed bool
}

// sweepEntry is the oracle's result for one sweepKey. oracleErr is non-empty
// when the gem raised for that specific (year, flag) cell; pairs is then empty.
type sweepEntry struct {
	pairs     []pair
	oracleErr string
}

// yearHolidaysRange makes ONE oracle round-trip for a whole region: every year
// in [fromYear, toYear] under every flag combo. It replaces the per-(year,flag)
// year_holidays calls the exhaustive sweep used to make (~324 -> 1 per region).
// A transport-level failure is returned as the error; per-cell Ruby errors ride
// along in each sweepEntry.oracleErr. Transport only: the set comparison and
// dedupe on the caller side are unchanged.
func (o *oracle) yearHolidaysRange(region string, fromYear, toYear int, combos []flagPair) (map[sweepKey]sweepEntry, error) {
	resp, err := o.call(request{
		Func:       "year_holidays_range",
		Region:     region,
		FromYear:   fromYear,
		ToYear:     toYear,
		FlagCombos: combos,
	})
	if err != nil {
		return nil, err
	}
	var rows []struct {
		Year     int    `json:"year"`
		Informal bool   `json:"informal"`
		Observed bool   `json:"observed"`
		OK       bool   `json:"ok"`
		Error    string `json:"error"`
		Holidays []pair `json:"holidays"`
	}
	if err := json.Unmarshal(resp.Result, &rows); err != nil {
		return nil, fmt.Errorf("decode year_holidays_range for %s: %w", region, err)
	}
	m := make(map[sweepKey]sweepEntry, len(rows))
	for _, r := range rows {
		sortPairs(r.Holidays)
		m[sweepKey{year: r.Year, informal: r.Informal, observed: r.Observed}] = sweepEntry{
			pairs:     r.Holidays,
			oracleErr: r.Error,
		}
	}
	return m, nil
}

// boolResult runs a request whose result is a bare boolean.
func (o *oracle) boolResult(req request) (bool, error) {
	resp, err := o.call(req)
	if err != nil {
		return false, err
	}
	var b bool
	if err := json.Unmarshal(resp.Result, &b); err != nil {
		return false, fmt.Errorf("decode bool for %s: %w", req.Func, err)
	}
	return b, nil
}

// stringList runs a request whose result is an array of strings (region codes).
func (o *oracle) stringList(req request) ([]string, error) {
	resp, err := o.call(req)
	if err != nil {
		return nil, err
	}
	var ss []string
	if err := json.Unmarshal(resp.Result, &ss); err != nil {
		return nil, fmt.Errorf("decode string list for %s: %w", req.Func, err)
	}
	return ss, nil
}

// loadCount runs a load_custom request and returns the oracle's reported count.
func (o *oracle) loadCount(req request) (int, error) {
	resp, err := o.call(req)
	if err != nil {
		return 0, err
	}
	var env struct {
		Loaded int `json:"loaded"`
	}
	if err := json.Unmarshal(resp.Result, &env); err != nil {
		return 0, fmt.Errorf("decode load_custom envelope: %w", err)
	}
	return env.Loaded, nil
}

// Close shuts the oracle down by closing its stdin (the main loop ends on EOF)
// and waiting for the process to exit.
func (o *oracle) Close() error {
	_ = o.stdin.Close()
	return o.cmd.Wait()
}

// ---- normalization -----------------------------------------------------------

// normalizeGo collapses a Go []holidays.Holiday into the same sorted (date,name)
// shape the oracle emits, so the two sides compare exactly.
func normalizeGo(hs []holidays.Holiday) []pair {
	pairs := make([]pair, 0, len(hs))
	for _, h := range hs {
		pairs = append(pairs, pair{
			Date: h.Date.UTC().Format("2006-01-02"),
			Name: h.Name,
		})
	}
	sortPairs(pairs)
	return pairs
}

// sortPairs orders by date ascending, then name ascending, matching the oracle.
func sortPairs(pairs []pair) {
	sort.SliceStable(pairs, func(i, j int) bool {
		if pairs[i].Date != pairs[j].Date {
			return pairs[i].Date < pairs[j].Date
		}
		return pairs[i].Name < pairs[j].Name
	})
}

// diffPairs returns the pairs present only in got (Go) and only in want (oracle).
// Used to print a readable mismatch report.
func diffPairs(got, want []pair) (onlyGo, onlyRuby []pair) {
	index := func(ps []pair) map[pair]int {
		m := make(map[pair]int, len(ps))
		for _, p := range ps {
			m[p]++
		}
		return m
	}
	gi, wi := index(got), index(want)
	for p, n := range gi {
		if extra := n - wi[p]; extra > 0 {
			for k := 0; k < extra; k++ {
				onlyGo = append(onlyGo, p)
			}
		}
	}
	for p, n := range wi {
		if extra := n - gi[p]; extra > 0 {
			for k := 0; k < extra; k++ {
				onlyRuby = append(onlyRuby, p)
			}
		}
	}
	sortPairs(onlyGo)
	sortPairs(onlyRuby)
	return onlyGo, onlyRuby
}

// pairsEqual reports whether two normalized lists are element-for-element equal.
func pairsEqual(a, b []pair) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// formatDiff renders a one-line-per-entry readable diff for t.Errorf output.
func formatDiff(onlyGo, onlyRuby []pair) string {
	var b strings.Builder
	b.WriteString("\n  only in Go (extra):")
	if len(onlyGo) == 0 {
		b.WriteString(" <none>")
	}
	for _, p := range onlyGo {
		fmt.Fprintf(&b, "\n    + %s  %s", p.Date, p.Name)
	}
	b.WriteString("\n  only in Ruby (missing from Go):")
	if len(onlyRuby) == 0 {
		b.WriteString(" <none>")
	}
	for _, p := range onlyRuby {
		fmt.Fprintf(&b, "\n    - %s  %s", p.Date, p.Name)
	}
	return b.String()
}

// ---- shared date helper ------------------------------------------------------

// mustDate parses a "YYYY-MM-DD" literal as a UTC calendar date. Corpus dates
// are all static literals, so a parse failure is a programmer error.
func mustDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(fmt.Sprintf("bad corpus date %q: %v", s, err))
	}
	return t.UTC()
}
