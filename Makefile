GO              ?= go
GOPATH          := $(shell $(GO) env GOPATH)
BIN             := bin
DEFS_REPO       := https://github.com/holidays/definitions.git
DEFS_TAG        ?= v9.2.0

.PHONY: build holidays gen-holidays vet staticcheck test generate update-definitions clean

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
# are never conflated into one check.
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

generate: gen-holidays
	$(BIN)/gen-holidays -in definitions -out internal/definitions

update-definitions:
	cd definitions && git fetch --tags && git checkout $(DEFS_TAG)
	@echo "Submodule now at $(DEFS_TAG). Don't forget: git add definitions && git commit"

clean:
	rm -rf $(BIN)
