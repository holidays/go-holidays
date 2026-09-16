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
})

var _ = Describe("HijriToGregorian", func() {
	It("converts a known Hijri date (1 Shawwal 1435) to its Gregorian equivalent", func() {
		// 1 Shawwal 1435 AH is Eid al-Fitr for Gregorian year 2014, i.e. the
		// pure calculation backing the (overridden) 2014 ramadan_feast.
		got := calc.HijriToGregorian(1435, 10, 1)
		Expect(got.Year()).To(Equal(2014))
	})
})
