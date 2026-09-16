package calc_test

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LunarToSolar", func() {
	Context("vn Giỗ tổ Hùng Vương (lunar month 3, day 10)", func() {
		// Regression coverage for go-holidays-ebh: the vietnameseLunarYearInfo
		// row for Gregorian year 1900+117 (2017) was missing a leap 6th month,
		// which shifted the cumulative day offset, and therefore every
		// computed Gregorian date, for every subsequent lunar year.
		DescribeTable("computes the correct Gregorian date",
			func(year, wantMonth, wantDay int) {
				got, err := calc.LunarToSolar(year, 3, 10, "vn")
				Expect(err).NotTo(HaveOccurred())
				Expect(got).To(Equal(time.Date(year, time.Month(wantMonth), wantDay, 0, 0, 0, 0, time.UTC)))
			},
			Entry("2018", 2018, 4, 25),
			Entry("2019", 2019, 4, 14),
			Entry("2020", 2020, 4, 2),
			Entry("2021", 2021, 4, 21),
			Entry("2022", 2022, 4, 10),
			Entry("2023", 2023, 4, 29),
		)
	})
})
