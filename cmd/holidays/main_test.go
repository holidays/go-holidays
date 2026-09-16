package main

import (
	"os"
	"time"

	holidays "github.com/holidays/go-holidays"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// jpBadYear is a year outside jpEquinox's supported range (1851-2150), used
// to trigger a real library error (rather than a CLI-level parse error) from
// holidays.On / YearHolidays / AnyHolidaysDuringWorkWeek when region "jp" is
// requested.
const jpBadYear = "1800-01-01"

var _ = Describe("reorderFlagsFirst", func() {
	It("moves a flag placed after positionals to the front", func() {
		got := reorderFlagsFirst([]string{"5", "2026-01-01", "--regions", "us"})
		Expect(got).To(Equal([]string{"--regions", "us", "--", "5", "2026-01-01"}))
	})

	It("keeps a flag placed before positionals in front", func() {
		got := reorderFlagsFirst([]string{"--regions", "us", "5", "2026-01-01"})
		Expect(got).To(Equal([]string{"--regions", "us", "--", "5", "2026-01-01"}))
	})

	It("keeps a bool flag after positionals without consuming the next token", func() {
		got := reorderFlagsFirst([]string{"2018-02-14", "--regions", "us", "--informal"})
		Expect(got).To(Equal([]string{"--regions", "us", "--informal", "--", "2018-02-14"}))
	})

	It("returns flags unchanged when there are no positionals", func() {
		got := reorderFlagsFirst([]string{"--regions", "us"})
		Expect(got).To(Equal([]string{"--regions", "us"}))
	})

	It("treats a bare negative integer as positional, not a flag", func() {
		got := reorderFlagsFirst([]string{"-3", "2026-01-01", "--regions", "us"})
		Expect(got).To(Equal([]string{"--regions", "us", "--", "-3", "2026-01-01"}))
	})
})

var _ = Describe("cmdRegions", func() {
	It("lists regions when given no arguments", func() {
		Expect(cmdRegions(nil)).To(Succeed())
	})

	It("rejects a trailing positional argument", func() {
		Expect(cmdRegions([]string{"extra-arg"})).To(MatchError("regions takes no arguments"))
	})

	It("rejects an unknown flag", func() {
		Expect(cmdRegions([]string{"--informal"})).To(HaveOccurred())
	})
})

var _ = Describe("cmdNext", func() {
	It("surfaces the library's count-must-be-positive error for a negative count", func() {
		err := cmdNext([]string{"-3", "2026-01-01", "--regions", "us"})
		Expect(err).To(MatchError("holidays.NextHolidays: count must be positive, got -3"))
	})

	It("still rejects a zero count with the library's error", func() {
		err := cmdNext([]string{"0", "2026-01-01", "--regions", "us"})
		Expect(err).To(MatchError("holidays.NextHolidays: count must be positive, got 0"))
	})

	It("still resolves a positive count with flags after the positionals", func() {
		err := cmdNext([]string{"5", "2026-01-01", "--regions", "us"})
		Expect(err).NotTo(HaveOccurred())
	})

	It("still resolves a positive count with flags before the positionals", func() {
		err := cmdNext([]string{"--regions", "us", "5", "2026-01-01"})
		Expect(err).NotTo(HaveOccurred())
	})

	It("rejects fewer than two positional arguments", func() {
		err := cmdNext([]string{"5"})
		Expect(err).To(MatchError("next expects 2 positional argument(s)"))
	})

	It("rejects a non-numeric count", func() {
		Expect(cmdNext([]string{"notanumber", "2026-01-01"})).To(HaveOccurred())
	})

	It("rejects a malformed date", func() {
		Expect(cmdNext([]string{"5", "bad-date"})).To(HaveOccurred())
	})
})

var _ = Describe("main", func() {
	var (
		savedArgs []string
		savedExit func(int)
	)

	BeforeEach(func() {
		savedArgs = os.Args
		savedExit = osExit
	})

	AfterEach(func() {
		os.Args = savedArgs
		osExit = savedExit
	})

	It("does not exit when the run succeeds", func() {
		os.Args = []string{"holidays", "regions"}
		called := false
		osExit = func(int) { called = true }
		main()
		Expect(called).To(BeFalse())
	})

	It("prints the error and exits with status 1 when the run fails", func() {
		os.Args = []string{"holidays"}
		var code int
		called := false
		osExit = func(c int) { called = true; code = c }
		main()
		Expect(called).To(BeTrue())
		Expect(code).To(Equal(1))
	})
})

var _ = Describe("run", func() {
	It("prints usage and errors when given no subcommand", func() {
		err := run(nil)
		Expect(err).To(MatchError("missing subcommand"))
	})

	It("dispatches on", func() {
		Expect(run([]string{"on", "2025-12-25", "--regions", "us"})).To(Succeed())
	})

	It("dispatches between", func() {
		Expect(run([]string{"between", "2025-01-01", "2025-01-02", "--regions", "us"})).To(Succeed())
	})

	It("dispatches year", func() {
		Expect(run([]string{"year", "2025", "--regions", "us"})).To(Succeed())
	})

	It("dispatches next", func() {
		Expect(run([]string{"next", "1", "2025-01-01", "--regions", "us"})).To(Succeed())
	})

	It("dispatches workweek", func() {
		Expect(run([]string{"workweek", "2025-01-01", "--regions", "us"})).To(Succeed())
	})

	It("dispatches regions", func() {
		Expect(run([]string{"regions"})).To(Succeed())
	})

	It("prints usage and returns nil for -h", func() {
		Expect(run([]string{"-h"})).To(Succeed())
	})

	It("prints usage and returns nil for --help", func() {
		Expect(run([]string{"--help"})).To(Succeed())
	})

	It("prints usage and returns nil for help", func() {
		Expect(run([]string{"help"})).To(Succeed())
	})

	It("prints usage and errors for an unknown subcommand", func() {
		err := run([]string{"bogus"})
		Expect(err).To(MatchError(`unknown subcommand "bogus"`))
	})
})

var _ = Describe("printUsage", func() {
	It("does not panic", func() {
		Expect(printUsage).NotTo(Panic())
	})
})

var _ = Describe("cmdOn", func() {
	It("resolves a valid date", func() {
		Expect(cmdOn([]string{"2025-12-25", "--regions", "us"})).To(Succeed())
	})

	It("rejects a missing positional argument", func() {
		Expect(cmdOn(nil)).To(MatchError("on expects 1 positional argument(s)"))
	})

	It("rejects a malformed date", func() {
		Expect(cmdOn([]string{"not-a-date"})).To(HaveOccurred())
	})

	It("surfaces a library error", func() {
		err := cmdOn([]string{jpBadYear, "--regions", "jp"})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("jp equinox"))
	})
})

var _ = Describe("cmdBetween", func() {
	It("resolves a valid range", func() {
		Expect(cmdBetween([]string{"2025-01-01", "2025-01-02", "--regions", "us"})).To(Succeed())
	})

	It("rejects fewer than two positional arguments", func() {
		err := cmdBetween([]string{"2025-01-01"})
		Expect(err).To(MatchError("between expects 2 positional argument(s)"))
	})

	It("rejects a malformed start date", func() {
		Expect(cmdBetween([]string{"bad", "2025-01-02"})).To(HaveOccurred())
	})

	It("rejects a malformed end date", func() {
		Expect(cmdBetween([]string{"2025-01-01", "bad"})).To(HaveOccurred())
	})

	It("surfaces the library's end-before-start error", func() {
		err := cmdBetween([]string{"2025-01-02", "2025-01-01"})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("is before start"))
	})
})

var _ = Describe("cmdYear", func() {
	It("resolves a valid year", func() {
		Expect(cmdYear([]string{"2025", "--regions", "us"})).To(Succeed())
	})

	It("rejects a missing positional argument", func() {
		Expect(cmdYear(nil)).To(MatchError("year expects 1 positional argument(s)"))
	})

	It("rejects a non-numeric year", func() {
		Expect(cmdYear([]string{"notanumber"})).To(HaveOccurred())
	})

	It("surfaces a library error", func() {
		err := cmdYear([]string{"1800", "--regions", "jp"})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("jp equinox"))
	})
})

var _ = Describe("cmdWorkweek", func() {
	It("resolves a valid date", func() {
		Expect(cmdWorkweek([]string{"2025-01-01", "--regions", "us"})).To(Succeed())
	})

	It("rejects a missing positional argument", func() {
		Expect(cmdWorkweek(nil)).To(MatchError("workweek expects 1 positional argument(s)"))
	})

	It("rejects a malformed date", func() {
		Expect(cmdWorkweek([]string{"bad"})).To(HaveOccurred())
	})

	It("surfaces a library error", func() {
		err := cmdWorkweek([]string{jpBadYear, "--regions", "jp"})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("jp equinox"))
	})
})

var _ = Describe("reorderFlagsFirst inline-value flags", func() {
	It("keeps a flag with an inline value without consuming the next token", func() {
		got := reorderFlagsFirst([]string{"--regions=us", "5", "2026-01-01"})
		Expect(got).To(Equal([]string{"--regions=us", "--", "5", "2026-01-01"}))
	})
})

var _ = Describe("isNegativeInt short inputs", func() {
	It("returns false for a lone dash", func() {
		Expect(isNegativeInt("-")).To(BeFalse())
	})

	It("returns false for an empty string", func() {
		Expect(isNegativeInt("")).To(BeFalse())
	})
})

var _ = Describe("printHolidays", func() {
	It("breaks ties between same-day holidays by name", func() {
		d := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		hs := []holidays.Holiday{
			{Date: d, Name: "Zeta", Regions: []string{"us"}},
			{Date: d, Name: "Alpha", Regions: []string{"us"}},
		}
		Expect(printHolidays(hs)).To(Succeed())
		Expect(hs[0].Name).To(Equal("Alpha"))
		Expect(hs[1].Name).To(Equal("Zeta"))
	})
})
