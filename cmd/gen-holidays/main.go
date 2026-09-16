// gen-holidays parses holidays/definitions YAML files and emits Go source under
// internal/definitions/, one <region>.go plus one <region>_test.go per country.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/holidays/go-holidays/internal/generator"
)

// osExit is os.Exit by default; tests swap it out so a failing run() can be
// observed without terminating the test binary.
var osExit = os.Exit

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "gen-holidays:", err)
		osExit(1)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet("gen-holidays", flag.ContinueOnError)
	inDir := fs.String("in", "definitions", "directory containing upstream YAML definitions")
	outDir := fs.String("out", "internal/definitions", "output directory for generated Go code")
	regionsArg := fs.String("regions", "", "comma-separated region list (empty == all available)")
	allowUnported := fs.Bool("allow-unported", false, "skip regions referencing methods that lack a Go implementation")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if err := generator.LoadSourceTag(*inDir); err != nil {
		return err
	}

	files, err := selectInputFiles(*inDir, *regionsArg)
	if err != nil {
		return err
	}
	if len(files) == 0 {
		return fmt.Errorf("no input YAML files matched")
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(*outDir, "helpers_test.go"), generator.EmitTestHelpers(), 0o644); err != nil {
		return fmt.Errorf("write helpers_test.go: %w", err)
	}

	var (
		generated []string
		skipped   []string
	)
	for _, path := range files {
		country := strings.TrimSuffix(filepath.Base(path), ".yaml")
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		rf, err := generator.ParseRegionFile(country, data)
		if err != nil {
			return err
		}
		if len(rf.Rules) == 0 {
			fmt.Fprintf(os.Stderr, "  skip %s (no months: block)\n", country)
			continue
		}
		if vErr := generator.Validate(rf); vErr != nil {
			var ve *generator.ValidationError
			if errors.As(vErr, &ve) && *allowUnported {
				skipped = append(skipped, fmt.Sprintf("%s: missing %s", ve.Country, strings.Join(ve.Missing, ", ")))
				continue
			}
			return vErr
		}

		defSrc, err := generator.EmitDefinitions(rf)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(*outDir, country+".go"), defSrc, 0o644); err != nil {
			return err
		}
		testPath := filepath.Join(*outDir, country+"_test.go")
		if len(rf.Tests) > 0 {
			// EmitTests, unlike EmitDefinitions, never fails: every value it
			// writes into the buffer (country name, table names, dates,
			// region codes) passes through strconv.Quote before landing in
			// the output, so it can only ever produce a valid Go string
			// literal, never syntax invalid enough for go/format to reject.
			// EmitDefinitions differs because it emits the country name as
			// a raw Go identifier (var <country>Rules = ...), which is
			// where invalid-identifier failures actually surface.
			testSrc, _ := generator.EmitTests(rf)
			if err := os.WriteFile(testPath, testSrc, 0o644); err != nil {
				return err
			}
		} else {
			_ = os.Remove(testPath)
		}
		generated = append(generated, country)
	}

	fmt.Printf("generated %d regions: %s\n", len(generated), strings.Join(generated, ", "))
	if len(skipped) > 0 {
		fmt.Println("skipped (unported methods):")
		for _, s := range skipped {
			fmt.Println("  -", s)
		}
	}
	return nil
}

func selectInputFiles(inDir, regionsArg string) ([]string, error) {
	if regionsArg != "" {
		var out []string
		for _, r := range strings.Split(regionsArg, ",") {
			r = strings.TrimSpace(r)
			if r == "" {
				continue
			}
			path := filepath.Join(inDir, r+".yaml")
			if _, err := os.Stat(path); err != nil {
				return nil, fmt.Errorf("region %s: %w", r, err)
			}
			out = append(out, path)
		}
		return out, nil
	}
	entries, err := os.ReadDir(inDir)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".yaml") {
			continue
		}
		if name == "index.yaml" || name == "METHODS.yml" {
			continue
		}
		out = append(out, filepath.Join(inDir, name))
	}
	sort.Strings(out)
	return out, nil
}
