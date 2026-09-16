GO              ?= go
GOPATH          := $(shell $(GO) env GOPATH)
BIN             := bin
DEFS_REPO       := https://github.com/holidays/definitions.git
DEFS_TAG        ?= v9.0.0

.PHONY: build holidays gen-holidays vet staticcheck test cover parity generate update-definitions clean

build: holidays gen-holidays

holidays:
	$(GO) build -o $(BIN)/holidays ./cmd/holidays

gen-holidays:
	$(GO) build -o $(BIN)/gen-holidays ./cmd/gen-holidays

vet:
	$(GO) vet ./...

staticcheck:
	@command -v staticcheck >/dev/null 2>&1 || $(GO) install honnef.co/go/tools/cmd/staticcheck@latest
	$(GOPATH)/bin/staticcheck ./...

# test runs go vet, then every package's tests with a per-package coverage
# profile written to <pkg-dir>/.cover.profile, then requires every function
# in every package (per `go tool cover -func`) to be at 100% coverage. Two
# passes: the first exits non-zero on any `go test` FAIL line (actual test
# failures, or a package with no test files at all -- treated the same way,
# by synthesizing a FAIL line when no profile got written); the second exits
# non-zero on any function below 100% (coverage gaps). The two failure modes
# are never conflated into one check. parity/ is excluded automatically:
# every file in it is gated behind the `parity` build tag, so `go list ./...`
# never lists it under the default build context.
test: vet
	@go list -f '{{.Dir}}/.cover.profile {{.ImportPath}}' ./... \
		| while read coverage package ; do \
			$(GO) test -coverprofile "$$coverage" "$$package"; \
			[ -f "$$coverage" ] || echo "FAIL	$$package has no test files"; \
		  done \
		| awk '{ print } /^FAIL/ { failures++ } END { exit failures }'
	@go list -f '{{.Dir}}/.cover.profile' ./... \
		| while read coverage ; do [ -f "$$coverage" ] && go tool cover -func "$$coverage" ; done \
		| awk '$$3 !~ /^100/ { print; gaps++ } END { exit gaps }'

# cover opens an HTML coverage report for one package's last `make test` run
# (use PKG=<package dir relative to repo root>, e.g. `make cover PKG=internal/calc`
# or `make cover PKG=.` for the root package).
cover:
	go tool cover -html="$(PKG)/.cover.profile"

# parity runs the Ruby<->Go comparison suite (build-tagged, excluded from `test`).
# Requires Ruby and the `holidays` gem installed, plus the `definitions/` submodule
# checked out (the oracle loads our region YAML into the gem via load_custom).
# See parity/README.md for the design and prerequisites.
parity:
	@ls definitions/*.yaml >/dev/null 2>&1 || { \
		echo "error: definitions/ submodule is empty; the oracle would load no region YAML and every region would mismatch." >&2; \
		echo "Run: git submodule update --init definitions" >&2; \
		exit 1; \
	}
	$(GO) test -tags parity -v -timeout=20m ./parity/...

generate: gen-holidays
	$(BIN)/gen-holidays -in definitions -out internal/definitions

update-definitions:
	cd definitions && git fetch --tags && git checkout $(DEFS_TAG)
	@echo "Submodule now at $(DEFS_TAG). Don't forget: git add definitions && git commit"

clean:
	rm -rf $(BIN)
