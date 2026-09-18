package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Coverage for holidays#392: Morocco's eid_al_fitr and eid_al_adha are
// derived from the arithmetic Hijri calendar (calc.HijriYearOccurrence),
// with a small Ministry of Habous and Islamic Affairs proclamation override
// table layered on top, matching definitions PR #391 ("ma-hijri-date-calculator")
// and the merged Ruby gem PR holidays/holidays#510.
var _ = Describe("ma feast methods", func() {
	var (
		eidAlFitr Method
		eidAlAdha Method
	)

	BeforeEach(func() {
		var ok bool
		eidAlFitr, ok = LookupMethod("eid_al_fitr")
		Expect(ok).To(BeTrue(), "eid_al_fitr must be registered")
		eidAlAdha, ok = LookupMethod("eid_al_adha")
		Expect(ok).To(BeTrue(), "eid_al_adha must be registered")
	})

	call := func(fn Method, year int) time.Time {
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Context("a year with a Ministry override", func() {
		It("eid_al_fitr returns the proclaimed date, not the pure arithmetic calculation", func() {
			// Arithmetic calendar gives 2022-05-03; the Ministry proclaimed 2022-05-02.
			Expect(call(eidAlFitr, 2022)).To(Equal(time.Date(2022, time.May, 2, 0, 0, 0, 0, time.UTC)))
		})

		It("eid_al_adha returns the proclaimed date", func() {
			// Arithmetic calendar gives 2021-07-20; the Ministry proclaimed 2021-07-21.
			Expect(call(eidAlAdha, 2021)).To(Equal(time.Date(2021, time.July, 21, 0, 0, 0, 0, time.UTC)))
		})
	})

	Context("a year with no override", func() {
		It("eid_al_fitr falls back to the Hijri calculation", func() {
			Expect(call(eidAlFitr, 2018)).To(Equal(time.Date(2018, time.June, 15, 0, 0, 0, 0, time.UTC)))
			Expect(call(eidAlFitr, 2023)).To(Equal(time.Date(2023, time.April, 22, 0, 0, 0, 0, time.UTC)))
		})

		It("eid_al_adha falls back to the Hijri calculation", func() {
			Expect(call(eidAlAdha, 2018)).To(Equal(time.Date(2018, time.August, 22, 0, 0, 0, 0, time.UTC)))
			Expect(call(eidAlAdha, 2023)).To(Equal(time.Date(2023, time.June, 29, 0, 0, 0, 0, time.UTC)))
		})
	})

	Context("a year the arithmetic calendar's search window misses entirely (1970-1973 for eid_al_adha)", func() {
		It("returns the zero time with no error, not a hard failure", func() {
			for _, year := range []int{1970, 1971, 1972, 1973} {
				got, err := eidAlAdha(MethodArgs{Year: year})
				Expect(err).NotTo(HaveOccurred(), "year %d", year)
				Expect(got.IsZero()).To(BeTrue(), "expected zero time for year %d, got %v", year, got)
			}
		})
	})
})
