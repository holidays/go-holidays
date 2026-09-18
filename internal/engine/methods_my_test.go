package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Coverage for go-holidays-jw1: hari_raya_puasa (1 Shawwal) and hari_raya_haji
// (10 Dhul-Hijjah) are derived from the arithmetic Islamic calendar
// (calc.HijriYearOccurrence), with a small gazette-proclamation override
// table layered on top, mirroring tr's ramadan_feast/sacrifice_feast (see
// methods_tr.go). Sourced from holidays/holidays#392 and the merged Ruby
// Holidays::Definition::CustomMethods::MY module.
var _ = Describe("my feast methods", func() {
	var (
		hariRayaPuasa Method
		hariRayaHaji  Method
	)

	BeforeEach(func() {
		var ok bool
		hariRayaPuasa, ok = LookupMethod("hari_raya_puasa")
		Expect(ok).To(BeTrue(), "hari_raya_puasa must be registered")
		hariRayaHaji, ok = LookupMethod("hari_raya_haji")
		Expect(ok).To(BeTrue(), "hari_raya_haji must be registered")
	})

	call := func(fn Method, year int) time.Time {
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Context("a year with a gazette override", func() {
		It("hari_raya_puasa returns the gazetted date, not the pure calculation", func() {
			Expect(call(hariRayaPuasa, 2014)).To(Equal(time.Date(2014, time.July, 28, 0, 0, 0, 0, time.UTC)))
			Expect(call(hariRayaPuasa, 2026)).To(Equal(time.Date(2026, time.March, 21, 0, 0, 0, 0, time.UTC)))
		})

		It("hari_raya_haji returns the gazetted date", func() {
			Expect(call(hariRayaHaji, 2016)).To(Equal(time.Date(2016, time.September, 12, 0, 0, 0, 0, time.UTC)))
			Expect(call(hariRayaHaji, 2019)).To(Equal(time.Date(2019, time.August, 11, 0, 0, 0, 0, time.UTC)))
		})
	})

	Context("a year with no override", func() {
		It("hari_raya_puasa falls back to the Hijri calculation", func() {
			Expect(call(hariRayaPuasa, 2018)).To(Equal(time.Date(2018, time.June, 15, 0, 0, 0, 0, time.UTC)))
			Expect(call(hariRayaPuasa, 2023)).To(Equal(time.Date(2023, time.April, 22, 0, 0, 0, 0, time.UTC)))
		})

		It("hari_raya_haji falls back to the Hijri calculation", func() {
			Expect(call(hariRayaHaji, 2018)).To(Equal(time.Date(2018, time.August, 22, 0, 0, 0, 0, time.UTC)))
			Expect(call(hariRayaHaji, 2023)).To(Equal(time.Date(2023, time.June, 29, 0, 0, 0, 0, time.UTC)))
		})
	})

	Context("a year the arithmetic calendar's search window misses entirely", func() {
		It("hari_raya_haji returns the zero time with no error for 1970-1973 (same Dhul-Hijjah-10 gap as tr's sacrifice_feast)", func() {
			for _, year := range []int{1970, 1971, 1972, 1973} {
				got, err := hariRayaHaji(MethodArgs{Year: year})
				Expect(err).NotTo(HaveOccurred(), "year %d", year)
				Expect(got.IsZero()).To(BeTrue(), "expected zero time for year %d, got %v", year, got)
			}
		})

		It("hari_raya_puasa returns the zero time with no error for 1965-1967 (same Shawwal-1 gap pattern)", func() {
			for _, year := range []int{1965, 1966, 1967} {
				got, err := hariRayaPuasa(MethodArgs{Year: year})
				Expect(err).NotTo(HaveOccurred(), "year %d", year)
				Expect(got.IsZero()).To(BeTrue(), "expected zero time for year %d, got %v", year, got)
			}
		})
	})
})
