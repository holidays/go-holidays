package holidays_test

import (
	"errors"
	"time"

	holidays "github.com/holidays/go-holidays"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// The lunar tables cover lunar years 1900..2100. Queries whose own dates fall
// in that range must work even though Between also resolves the following
// year to catch holidays observed back across Dec 31.
var _ = Describe("lunar table range", func() {
	var (
		lunarRegions = []string{"cn", "hk", "kr", "vn", "sg"}
		day          = func(y, m, d int) time.Time { return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC) }
	)

	for _, region := range lunarRegions {
		opts := holidays.Options{Regions: []string{region}}
		for _, year := range []int{2049, 2100} {
			It("resolves "+region+" queries in the border year", func() {
				hs, err := holidays.YearHolidays(year, opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(hs).NotTo(BeEmpty())

				hs, err = holidays.Between(day(year, 1, 1), day(year, 12, 31), opts)
				Expect(err).NotTo(HaveOccurred())
				Expect(hs).NotTo(BeEmpty())

				_, err = holidays.YearHolidaysFrom(day(year, 1, 1), opts)
				Expect(err).NotTo(HaveOccurred())

				_, err = holidays.On(day(year, 12, 31), opts)
				Expect(err).NotTo(HaveOccurred())
			})
		}

		It("returns ErrUnsupportedYear for "+region+" queries past the tables", func() {
			_, err := holidays.YearHolidays(2101, opts)
			Expect(errors.Is(err, holidays.ErrUnsupportedYear)).To(BeTrue())

			_, err = holidays.Between(day(2101, 3, 1), day(2101, 3, 31), opts)
			Expect(errors.Is(err, holidays.ErrUnsupportedYear)).To(BeTrue())
		})
	}

	It("finds lunar holidays in 2100", func() {
		on := func(d time.Time, region string) []holidays.Holiday {
			hs, err := holidays.On(d, holidays.Options{Regions: []string{region}})
			Expect(err).NotTo(HaveOccurred())
			return hs
		}
		Expect(hasNamedHoliday(on(day(2100, 2, 9), "cn"), "春节")).To(BeTrue())
		Expect(hasNamedHoliday(on(day(2100, 4, 5), "cn"), "清明节")).To(BeTrue())
		Expect(hasNamedHoliday(on(day(2100, 9, 18), "kr"), "추석")).To(BeTrue())
		Expect(hasNamedHoliday(on(day(2100, 4, 18), "vn"), "Giỗ tổ Hùng Vương")).To(BeTrue())
		Expect(hasNamedHoliday(on(day(2057, 10, 6), "hk"), "Chung Yeung Festival")).To(BeTrue())
	})
})
