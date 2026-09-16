# Project Conventions

## Tech Stack

- Go 1.24.4 (`go.mod` module `github.com/holidays/go-holidays`)
- Runtime dependency: `gopkg.in/yaml.v3 v3.0.1` (used only by the code generator).
  Test-only dependencies: `github.com/onsi/ginkgo/v2` + `github.com/onsi/gomega`
  (the required test framework, see Testing section).
- Build/test orchestration: `Makefile` (`make build`, `make test`, `make generate`, `make update-definitions`)
- Lint: `go vet ./...` (required, part of `make test`) + optional `staticcheck`
- Upstream data: `definitions/` is a git submodule of `github.com/holidays/definitions` (the Ruby gem's YAML rules)

## What This Is

A Go port of the Ruby `holidays` gem. Upstream YAML region definitions are
compiled into Go source by a generator, then resolved at runtime by an engine.
All 79 upstream regions pass their generated tests.

## Architecture

Style: **layered, code-generation pipeline.** Two flows:

**Generate-time** (`make generate`): YAML → Go source
```
definitions/*.yaml  →  internal/generator (parse → validate → emit)  →  internal/definitions/*.go (+ *_test.go)
                       via cmd/gen-holidays
```

**Run-time**: public API → engine → registered rules + methods
```
holidays.go (public API at the module root: github.com/holidays/go-holidays —
             On/Between/YearHolidays/NextHolidays/..., plus a blank import of
             internal/definitions so every region registers itself)
  → internal/engine (ResolveYear: walk rules, filter by region/type/year, compute dates)
      → internal/definition (HolidayRule, YearRange types + matching logic)
      → internal/calc (date math: easter, lunar, day-of-month, observance)
      → registered Method funcs (builtins + per-country methods_*.go)
```

- **One import, no facade.** The public API lives at the module root
  (`package holidays`, path `github.com/holidays/go-holidays`) and holds the real
  logic, not forwarders. Everything it depends on is under `internal/`, so the
  root package is the entire public surface. This mirrors the Ruby gem, where
  `lib/holidays.rb` holds the API and generated per-region data sits in
  `lib/generated_definitions/`.
- **There is no region-free build.** The root package blank-imports
  `internal/definitions`, so the built-in regions are always registered. The gem
  has no region-free entry point either, and its `:any` branch (empty regions)
  eagerly loads every region, matching our empty `Options{}`. A `LoadCustom`-only
  caller names its own region codes rather than relying on an empty `Regions`.
- **Registration is via `init()` side effects.** Each `internal/definitions/xx.go`
  calls `engine.RegisterCountry("xx", xxRules)` in `init()`. Each
  `internal/engine/methods_*.go` and `builtins.go` calls
  `engine.RegisterMethod(name, fn)` in `init()`. Callers do nothing; importing the
  root package is enough.
- **Error handling:** Go idiom — return `(T, error)`, wrap with `fmt.Errorf("...: %w", err)`.
  `RegisterMethod` panics on duplicate name (programmer error, generate-time).
- **Concurrency:** registry maps guarded by `sync.RWMutex` (`methodMu`, `regionMu`, `cacheMu`).
- **Region matching** (`engine.ruleMatchesRequested`): exact, parent (`us` matches request `us_ga`),
  and wildcard (`gb_` matches `gb`, `gb_sct`, ...). Empty region list = all regions.
- **Dates are always UTC, truncated to day** (`time.Date(..., time.UTC)`, `truncateToDay`).

## Structure

```
holidays.go               public API (On, Between, YearHolidays, NextHolidays,
                          AnyHolidaysDuringWorkWeek) + the blank import of
                          internal/definitions that registers every region
holiday.go                Holiday result type
options.go                Options{Regions, Informal, Observed}
cache.go                  CacheBetween/ResetCache + range cache (cacheKey/cacheEntry)
load_custom.go            runtime YAML rule loading (LoadCustom)
ginkgo_deps.go            tools-tagged dependency anchor for Ginkgo/Gomega
cmd/
  holidays/main.go        CLI wrapper (subcommands: on/between/year/next/workweek/regions)
  gen-holidays/main.go    code generator entry point
internal/
  definition/types.go     HolidayRule, YearRange, HolidayType — pure data + matching
  engine/
    registry.go           RegisterMethod/RegisterCountry, MethodArgs, region matching
    resolve.go            ResolveYear, computeDate, nthOrLastWeekday, applyObserved
    builtins.go           the ~15 well-known methods (easter, to_monday_if_sunday, ...)
    methods_<cc>.go       per-country custom methods (us, jp, nz, ...)
  calc/                   pure date math (easter.go, lunar.go, day_of_month.go, observance.go)
  definitions/            GENERATED: <cc>.go (rules) + <cc>_test.go (table tests) + helpers_test.go
  generator/              parse.go (YAML→RegionFile), emit.go, emit_tests.go, Validate
definitions/              git submodule: upstream YAML (DO NOT hand-edit), VERSION.txt
```

## Naming

- Files: `snake_case.go` (`load_custom.go`, `methods_us.go`, `day_of_month.go`). Per-country
  files keyed by lowercase region code (`be_fr.go`, `rs_cyrl.go`).
- Exported funcs: Go `PascalCase` (`Between`, `ResolveYear`, `RegisterMethod`).
- Unexported helpers: `camelCase` (`truncateToDay`, `nthOrLastWeekday`, `ruleMatchesRequested`).
- Registered method names: `snake_case` strings matching upstream Ruby (`to_monday_if_sunday`,
  `day_after_thanksgiving`) — these are the YAML `function:`/`observed:` identifiers.
- Rule variables: `<cc>Rules` (e.g. `usRules`).

## Testing

- Framework: **Ginkgo v2 + Gomega are the required test framework**
  (`github.com/onsi/ginkgo/v2`, `github.com/onsi/gomega`). This reverses the old
  stdlib-only test rule; see `go-holidays-hk6` (epic) for the migration. The
  runtime stdlib-only constraint below still applies to non-test code — this
  reversal is test-scope only.
- BDD structure: `Describe`/`Context`/`It` for hand-written behavior tests,
  Gomega matchers (`Expect(...).To(...)`) instead of manual `if`/`t.Errorf`.
- Region tests are **generated table tests** (`internal/definitions/<cc>_test.go`), package
  `definitions_test`, header `// Code generated by gen-holidays ... DO NOT EDIT.`,
  emitted as Ginkgo `DescribeTable`/`Entry` (one `Entry` per date case) rather than
  a loop over a slice.
- Each test package needs exactly one Ginkgo bootstrap file wiring
  `RunSpecs(t, "...")` into `go test`; see the per-package bootstrap files added
  under epic `go-holidays-hk6`.
- Hand-written tests use internal/external packages: `cache_internal_test.go` (package
  `holidays`, white-box), `holidays_test.go` / `region_test.go` / `load_custom_test.go`
  (black-box, package `holidays_test`).
- Test naming: Ginkgo spec descriptions replace the old `TestUS_000_ShroveTuesday` /
  `TestXxx_Scenario` Go func-name convention; describe behavior in plain English
  inside `Describe`/`It` blocks instead.
- Always run via `make test` (runs `go vet` first, then `go test ./...` which
  runs the Ginkgo suites, then a gate requiring 100% statement coverage on
  every package).
- **Every package must be at 100% statement coverage.** `make test` fails if
  any package (per `go test -cover ./...`) is below 100%, including a package
  with no test files at all.

### Test taxonomy

Same categories as the Ruby gem (`holidays/holidays`, `test/e2e/README.md`), adapted
to Go's rule that white-box tests must sit in the package directory:

- **unit** — isolated, no real definition data. Lives next to the code it covers.
  White-box (`package <pkg>`) when it must reach unexported internals; the file is
  then named `*_internal_test.go`. Examples: `cache_internal_test.go` (beside
  `cache.go`, exercises `cacheStore`/`cacheFind`/`optionsKey` with sentinel
  `[]Holiday`), the `internal/calc`, `internal/engine`, and `internal/generator`
  suites.
- **integration** — controlled fixture YAML written at test time, stable across
  `definitions/` bumps. Black-box. Example: `load_custom_test.go`.
- **e2e** — end-user-visible behavior through the public API against the real
  pinned `definitions/`; intentionally coupled to that data, so a definitions bump
  can legitimately change what these assert. Black-box. Examples: `holidays_test.go`,
  `region_test.go`, and the generated `internal/definitions/<cc>_test.go` tables.

`cache_internal_test.go` is white-box and lives at the module root only because
`cache.go` does; by this taxonomy it is a unit test sitting next to its code, not an
e2e test that wandered out of place. A test with no dependency on real definition
content belongs in the unit or integration category, never e2e.

## Do NOT

- **Do NOT hand-edit `internal/definitions/*.go` or `*_test.go`** — they are generated.
  Change the YAML (upstream) or the generator/methods, then `make generate`.
- **Do NOT hand-edit `definitions/*.yaml`** — it's a submodule; bump via `make update-definitions`.
- Do NOT add new runtime dependencies — stdlib-only at runtime is a deliberate constraint
  (yaml.v3 is generator-only). This does not apply to tests: Ginkgo/Gomega are the
  required test framework (see Testing section above).
- Do NOT register a method name twice (`RegisterMethod` panics).
- Do NOT introduce a Repository/interface wrapper over the registry — registration is the pattern.
- Do NOT use local time — all date construction is `time.UTC`.

## Definitions version

Target: **the latest upstream definitions tag** (currently **v9.0.0**, what the
`definitions/` submodule is pinned at). The parity oracle gem is pinned in
`parity/Gemfile` (currently `holidays` 11.5.0).

`generator.SourceTag` is read from the submodule's `VERSION.txt` at generate
time (`internal/generator/source_tag.go`), so the generated file headers can never
disagree with the pinned submodule. `Makefile DEFS_TAG` is the checkout target
for `make update-definitions` and is bumped alongside the submodule.

The compiled rules in `internal/definitions/*.go` match the pinned submodule:
regenerating produces identical output, all 81 regions pass (2673/2673 tests),
and the exhaustive parity sweep is clean with an empty `knownDivergences`.

The parity sweep now covers every region including `jp` (290 serviceable, 0
oracle-unsupported). `jp` was previously skipped because the gem's ported
`jp_next_weekday` reaches for a `Holidays::JP` module that only exists when the
gem loads its own vendored `jp` file; a shim in `parity/oracle.rb` rebuilds that
module from our own loaded rules. `isOracleUnsupported` and its per-case guards
stay as a harmless safety net that no region hits today.

v9.0.0 is a breaking upstream change: every `methods:` / `ruby:` block was
removed from the region YAML and the method bodies now live in the consuming
library. For Go that means each referenced `function:` must be registered in
`internal/engine` (the generator's `Validate` enforces this). New in this bump:
`cn` (China, `cn_qingming` plus the shared Chinese lunar table) and
`kr_seollal_eve`.

## Runtime loading vs. pre-loading (current priority)

Two ways rules enter the registry:

- **Pre-loading (compile-time):** `internal/definitions/*.go`, generated from YAML and
  registered via `init()`. This is the current default for built-in regions.
- **Runtime loading:** `holidays.LoadCustom(paths...)` parses YAML at runtime and
  registers rules under `custom:<basename>` keys; `UnloadCustom` removes them.
  `function:`/`observed:` references must be pre-registered via `holidays.RegisterMethod`.

**Runtime loading is the prioritized path for current work.** Treat `LoadCustom` /
`RegisterMethod` / `UnloadCustom` as the primary surface. Pre-loading (the generated
compile-time path) stays as-is for now; reworking or consolidating it is deferred to
a future, separately-scoped discussion — do not change the generation pipeline beyond
the version-label fix above without raising it first.

**API stability decision (2026-09-14, ahead of the v1.0.0 tag):** there is no concrete
plan to consolidate pre-loading and runtime loading, so a future consolidation, if it
ever happens, is an internal implementation detail only. `LoadCustom(paths ...string) error`,
`UnloadCustom(paths ...string)`, `RegisterMethod(name string, fn func(MethodArgs) (time.Time, error))`,
and `MethodArgs` are locked as of v1.0.0: only additive changes (new functions, new
optional fields) are allowed going forward, never a signature change or removal.
