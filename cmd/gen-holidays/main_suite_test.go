package main

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// TestGenHolidaysCLI is the single Ginkgo bootstrap for the cmd/gen-holidays
// test binary. It is white-box (package main) so specs can exercise
// unexported helpers such as run and selectInputFiles directly.
func TestGenHolidaysCLI(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cmd/gen-holidays suite")
}
