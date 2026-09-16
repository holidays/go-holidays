package calc_test

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HijriYearOccurrence", func() {
	DescribeTable("finds the Gregorian occurrence of Shawwal 1 (Ramazan Bayramı) within a Gregorian year",
		func(gregorianYear, month, day int) {
			got, ok := calc.HijriYearOccurrence(gregorianYear, 10, 1)
			Expect(ok).To(BeTrue())
			Expect(got).To(Equal(time.Date(gregorianYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)))
		},
		// Pure arithmetic-calendar years (no Diyanet proclamation override),
		// matching the definitions tr.yaml test corpus.
		Entry("2031", 2031, 1, 25),
		Entry("2032", 2032, 1, 14),
	)

	DescribeTable("finds the Gregorian occurrence of Dhu al-Hijjah 10 (Kurban Bayramı) within a Gregorian year",
		func(gregorianYear, month, day int) {
			got, ok := calc.HijriYearOccurrence(gregorianYear, 12, 10)
			Expect(ok).To(BeTrue())
			Expect(got).To(Equal(time.Date(gregorianYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)))
		},
		Entry("2031", 2031, 4, 3),
		Entry("2032", 2032, 3, 22),
	)

	// Roughly once every 33 years a fixed Hijri date drifts across a
	// Gregorian year boundary fast enough that the function's +/-1-year
	// search window misses it entirely; 1970-1973 is one such gap for Dhu
	// al-Hijjah 10 (see the trFeastDate comment in methods_tr.go).
	DescribeTable("reports no occurrence when the search window misses the date",
		func(gregorianYear int) {
			_, ok := calc.HijriYearOccurrence(gregorianYear, 12, 10)
			Expect(ok).To(BeFalse())
		},
		Entry("1970", 1970),
		Entry("1971", 1971),
		Entry("1972", 1972),
		Entry("1973", 1973),
	)
})

var _ = Describe("HijriToGregorian", func() {
	It("converts a known Hijri date (1 Shawwal 1435) to its Gregorian equivalent", func() {
		// 1 Shawwal 1435 AH is Eid al-Fitr for Gregorian year 2014, i.e. the
		// pure calculation backing the (overridden) 2014 ramadan_feast.
		got := calc.HijriToGregorian(1435, 10, 1)
		Expect(got.Year()).To(Equal(2014))
	})
})
