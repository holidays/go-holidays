GO              ?= go
GOPATH          := $(shell $(GO) env GOPATH)
BIN             := bin
DEFS_REPO       := https://github.com/holidays/definitions.git
DEFS_TAG        ?= v9.0.0
COVERAGE_MIN    := 100.0

.PHONY: build holidays gen-holidays vet staticcheck test coverage-check parity generate update-definitions clean

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

test: vet
	$(GO) test ./...

# coverage-check enforces COVERAGE_MIN% statement coverage on every package
# under `go list ./...` (parity/ is excluded automatically: it's gated behind
# the `parity` build tag, so plain `go test ./...` never builds it). Reads
# `go test -cover`'s own per-package summary line rather than a coverprofile.
# A package with no test files does NOT print "ok" + a coverage line the way
# a tested package does (which would otherwise let it slide through silently)
# -- with -cover it instead prints a line with an empty status field and
# "coverage: 0.0% of statements", which the threshold check below catches on
# its own. The literal "[no test files]" / "[no statements]" markers are kept
# as a defensive backstop in case that shape ever appears (older Go, or a
# package with truly zero testable statements). Package names are pulled out
# by matching the module import-path pattern rather than a fixed field index,
# since the no-test-files line shifts every field left by one.
coverage-check:
	@output="$$($(GO) test -cover ./... 2>&1)"; \
	status=$$?; \
	echo "$$output"; \
	echo "$$output" | awk -v min=$(COVERAGE_MIN) ' \
		{ \
			pkg = $$0; \
			if (match(pkg, /github\.com\/holidays\/go-holidays[^ \t]*/)) { pkg = substr(pkg, RSTART, RLENGTH) } \
			else { pkg = $$0 } \
		} \
		/\[no test files\]/ { printf "FAIL: %s has no test files\n", pkg; bad=1; next } \
		/\[no statements\]/ { printf "FAIL: %s has no statements to cover\n", pkg; bad=1; next } \
		/coverage: [0-9.]+% of statements/ { \
			pct = $$0; \
			sub(/.*coverage: /, "", pct); \
			sub(/% of statements.*/, "", pct); \
			if (pct + 0 < min + 0) { printf "FAIL: %s coverage %s%% < %s%%\n", pkg, pct, min; bad=1 } \
			next \
		} \
		END { exit bad }'; \
	awkstatus=$$?; \
	if [ $$status -ne 0 ]; then echo "coverage-check: go test failed"; exit $$status; fi; \
	if [ $$awkstatus -ne 0 ]; then echo "coverage-check: one or more packages below $(COVERAGE_MIN)% coverage"; exit 1; fi; \
	echo "coverage-check: all packages at >= $(COVERAGE_MIN)% coverage"

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
