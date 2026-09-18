package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Regression coverage for go-holidays-2nv: ramadan_feast and sacrifice_feast
// used to be static per-year lookup tables covering only 2014-2030. They are
// now a small Diyanet-proclamation override table layered on top of an
// arithmetic Hijri calendar calculation (calc.HijriYearOccurrence), matching
// definitions PR #385 ("tr-hijri-date-calculator"). This extends coverage to
// any year, not just a hard-coded range.
var _ = Describe("tr feast methods", func() {
	var (
		ramadanFeast   Method
		sacrificeFeast Method
	)

	BeforeEach(func() {
		var ok bool
		ramadanFeast, ok = LookupMethod("ramadan_feast")
		Expect(ok).To(BeTrue(), "ramadan_feast must be registered")
		sacrificeFeast, ok = LookupMethod("sacrifice_feast")
		Expect(ok).To(BeTrue(), "sacrifice_feast must be registered")
	})

	call := func(fn Method, year int) time.Time {
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Context("a year with a Diyanet override", func() {
		It("ramadan_feast returns the proclaimed date, not the pure calculation", func() {
			Expect(call(ramadanFeast, 2014)).To(Equal(time.Date(2014, time.July, 28, 0, 0, 0, 0, time.UTC)))
		})

		It("sacrifice_feast returns the proclaimed date", func() {
			Expect(call(sacrificeFeast, 2014)).To(Equal(time.Date(2014, time.October, 4, 0, 0, 0, 0, time.UTC)))
		})
	})

	Context("years past the old static table (2031, 2032), with no override", func() {
		It("ramadan_feast falls back to the Hijri calculation", func() {
			Expect(call(ramadanFeast, 2031)).To(Equal(time.Date(2031, time.January, 25, 0, 0, 0, 0, time.UTC)))
			Expect(call(ramadanFeast, 2032)).To(Equal(time.Date(2032, time.January, 14, 0, 0, 0, 0, time.UTC)))
		})

		It("sacrifice_feast falls back to the Hijri calculation", func() {
			Expect(call(sacrificeFeast, 2031)).To(Equal(time.Date(2031, time.April, 3, 0, 0, 0, 0, time.UTC)))
			Expect(call(sacrificeFeast, 2032)).To(Equal(time.Date(2032, time.March, 22, 0, 0, 0, 0, time.UTC)))
		})
	})

	Context("a year within the old table's range but without an override (e.g. 2018)", func() {
		It("ramadan_feast still matches the previously hard-coded value", func() {
			Expect(call(ramadanFeast, 2018)).To(Equal(time.Date(2018, time.June, 15, 0, 0, 0, 0, time.UTC)))
		})
	})

	Context("a year the arithmetic calendar's search window misses entirely (e.g. 1970-1973 for sacrifice_feast)", func() {
		It("returns the zero time with no error, not a hard failure", func() {
			for _, year := range []int{1970, 1971, 1972, 1973} {
				got, err := sacrificeFeast(MethodArgs{Year: year})
				Expect(err).NotTo(HaveOccurred(), "year %d", year)
				Expect(got.IsZero()).To(BeTrue(), "expected zero time for year %d, got %v", year, got)
			}
		})
	})
})
