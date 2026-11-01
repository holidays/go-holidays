package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// versionFile is the file inside the definitions checkout that records the
// upstream release the data came from.
const versionFile = "VERSION.txt"

// sourceTagFormat is what a normalized tag must look like. The tag is
// written unquoted into a generated header comment, so a newline or other
// stray text in VERSION.txt must not get through.
var sourceTagFormat = regexp.MustCompile(`^v[0-9A-Za-z][0-9A-Za-z.+-]*$`)

// SourceTag identifies the upstream version this code was generated from.
// Embedded into the file header for traceability. It carries no default:
// gen-holidays sets it via LoadSourceTag so the headers can never disagree
// with the definitions checkout they were generated from.
var SourceTag string

// LoadSourceTag reads <inDir>/VERSION.txt and sets SourceTag from it.
// A missing, unreadable, or blank VERSION.txt is an error, so generation
// fails loudly instead of emitting an empty or stale tag.
func LoadSourceTag(inDir string) error {
	path := filepath.Join(inDir, versionFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	tag, err := NormalizeSourceTag(string(data))
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	SourceTag = tag
	return nil
}

// NormalizeSourceTag turns raw VERSION.txt contents into the tag form used in
// generated headers: surrounding whitespace trimmed, "v" prefixed when absent.
func NormalizeSourceTag(raw string) (string, error) {
	tag := strings.TrimSpace(raw)
	if tag == "" {
		return "", fmt.Errorf("%s is blank", versionFile)
	}
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}
	if !sourceTagFormat.MatchString(tag) {
		return "", fmt.Errorf("%s is not a version: %q", versionFile, tag)
	}
	return tag, nil
}
