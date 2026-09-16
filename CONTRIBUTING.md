# How to contribute

There are multiple ways to help. We rely on the upstream
[holidays/definitions](https://github.com/holidays/definitions) project and
its contributors to keep holiday data accurate and up to date, and pull
requests to this repository to address bugs or implement new features are
always welcome.

## Code of Conduct

Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.
Everyone interacting with this project is expected to abide by its terms.

## AI Usage

Please read our [AI Usage Policy](AI_POLICY.md) before contributing.

## Commit requirements

All commits must be GPG-signed and include a `Signed-off-by` trailer. Use both
the `-S` and `-s` flags together:

```sh
git commit -S -s -m "Your commit message"
```

**GPG signing** (`-S`) verifies that the commit genuinely came from you. If you
haven't set up GPG signing with Git yet, GitHub has a guide:
https://docs.github.com/en/authentication/managing-commit-signature-verification

**Signed-off-by** (`-s`) is your acknowledgment of the
[Developer Certificate of Origin](https://developercertificate.org/), certifying
that you have the right to submit the contribution under this project's license.
It appends the following to your commit message automatically:

```
Signed-off-by: Your Name <your@email.com>
```

## General note on the definitions submodule

Definitions live in a git submodule. Clone with:

```sh
git clone --recurse-submodules https://github.com/holidays/go-holidays
```

or, in an existing clone:

```sh
git submodule update --init
```

To bump the submodule to a newer upstream tag:

```sh
make update-definitions DEFS_TAG=vX.Y.Z
git add definitions
git commit -s -m "vendor holidays/definitions vX.Y.Z"
make generate
```

## For definition updates

Definition changes belong upstream, not here. Definitions are written in YAML
and live in the [holidays/definitions](https://github.com/holidays/definitions)
repository so they can be used by tools written in other languages; a
complete guide to the format is in its
[SYNTAX guide](https://github.com/holidays/definitions/blob/master/doc/SYNTAX.md).

Once you have an idea of what you want to change, see that repository's own
[CONTRIBUTING guide](https://github.com/holidays/definitions/blob/master/doc/CONTRIBUTING.md).
After that PR is accepted upstream, this repository's maintainers are
responsible for bumping the submodule, regenerating, and releasing.

Never hand-edit `definitions/*.yaml` (it is a submodule checkout) or
`internal/definitions/*.go` / `*_test.go` (they are generated).

## For non-definition functionality

* Fork this repository.
* Make your changes. Run `make test` to execute the test suite (this also
  enforces 100% statement coverage on every function, so new code needs
  tests).
* Open a PR pointing back to `main`.

## Regenerating definitions

The generated `internal/definitions/*.go` files are checked in. Regenerate
after bumping the submodule or after adding new per-region methods:

```sh
make generate                       # all regions, fail on unported methods
bin/gen-holidays -allow-unported    # all regions, skip those with missing methods
bin/gen-holidays -regions us,gb     # only specific regions
```

The generator hard-fails on any `function:` or `observed:` YAML reference that
lacks a registered Go implementation.

### Adding a missing method

1. Write the Go function in `internal/engine/methods_<cc>.go`.
2. Register it in that file's `init()` with
   `engine.RegisterMethod("<name>", func(a MethodArgs) (time.Time, error) { ... })`.
3. Re-run `make generate`.

## Testing

```sh
make test          # go vet + per-package tests + 100% per-function coverage gate
```

Every function in every package must be at 100% statement coverage, not just
each package's overall average. `make test` writes a coverage profile per
package to `<pkg-dir>/.cover.profile`, then checks every line of `go tool
cover -func`'s output (one per function) for that package - it fails if any
function is below 100%, or if a package has no test files at all. Drill into
one package's report with `make cover PKG=<pkg-dir>` (e.g.
`make cover PKG=internal/calc`), which opens an HTML view of its last
`.cover.profile`. Ginkgo v2 and Gomega are the required test framework.
Region tests under `internal/definitions` are generated table tests and must
not be hand-edited; change the upstream YAML or the generator, then `make
generate`.

### Test taxonomy

The same categories the Ruby gem uses (see
[`holidays/holidays` `test/e2e/README.md`](https://github.com/holidays/holidays/blob/master/test/e2e/README.md)),
adapted to Go's rule that a white-box test must live in the directory of the
package it tests:

* **unit** - isolated, no dependency on real definition data. Lives next to the
  code it covers. White-box (`package <pkg>`, file named `*_internal_test.go`)
  only when it must reach unexported internals. Examples: `cache_internal_test.go`
  beside `cache.go`; the `internal/calc`, `internal/engine`, and
  `internal/generator` suites.
* **integration** - uses controlled fixture YAML created at test time, so it stays
  stable across `definitions/` bumps. Black-box (`package holidays_test`). Example:
  `load_custom_test.go`.
* **e2e** - end-user-visible behavior through the public API against the real
  pinned `definitions/`. Intentionally coupled to that data: a definitions bump
  can legitimately change what these assert. Black-box. Examples:
  `holidays_test.go`, `region_test.go`, and the generated
  `internal/definitions/<cc>_test.go` tables.

A test with no dependency on real definition content belongs in the unit or
integration category, never e2e.

## Parity suite

`parity/` compares this library's output against the Ruby gem pinned in
`parity/Gemfile`, loading our own `definitions/` YAML into the gem so both
sides resolve the same rules. It is run by the required "parity" CI check on
every PR; do not run it locally, it needs Ruby and the pinned gem installed.
See `parity/README.md` for the design.

## Local development helpers

* `make build` - builds `bin/holidays` and `bin/gen-holidays`
* `make test` - runs `go vet`, the full test suite, and a 100% per-function coverage gate
* `make cover PKG=<pkg-dir>` - opens an HTML coverage report for one package's last `make test` run
* `make vet` - runs `go vet ./...` only
* `make staticcheck` - runs `staticcheck ./...` (installs it if missing)
* `make generate` - regenerates `internal/definitions` from the YAML submodule
* `make update-definitions DEFS_TAG=vX.Y.Z` - checks the submodule out at a tag
* `make parity` - runs the Ruby-vs-Go comparison suite (CI only, see above)
* `make clean` - removes `bin/`
