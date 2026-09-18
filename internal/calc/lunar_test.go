package calc_test

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LunarToSolar", func() {
	DescribeTable("converts a lunar date to its Gregorian equivalent",
		func(year, month, day int, region string, wantYear, wantMonth, wantDay int) {
			got, err := calc.LunarToSolar(year, month, day, region)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(time.Date(wantYear, time.Month(wantMonth), wantDay, 0, 0, 0, 0, time.UTC)))
		},
		// Cross-checked against the oracle-verified internal/definitions
		// generated table tests, which exercise lunar_to_solar for these
		// exact (year, month, day, region) inputs.
		Entry("kr Seollal 2016 (Korean table)", 2016, 1, 1, "kr", 2016, 2, 8),
		Entry("hk Lunar New Year 2016 (Chinese table)", 2016, 1, 1, "hk", 2016, 2, 8),
		Entry("cn Chinese New Year 2025 (Chinese table, high year index)", 2025, 1, 1, "cn", 2025, 1, 29),
		Entry("sg Lunar New Year 2019 (Chinese table, shared region alias)", 2019, 1, 1, "sg", 2019, 2, 5),
		Entry("vn Giỗ tổ Hùng Vương 2017 (Vietnamese table, non-1st month)", 2017, 3, 10, "vn", 2017, 4, 6),
		// First and last rows of the lookup tables (1900 and 2049).
		Entry("earliest supported year (yearDiff == 0)", 1900, 1, 1, "cn", 1900, 1, 31),
		Entry("latest supported year (yearDiff == len-1)", 2049, 1, 1, "cn", 2049, 2, 2),
	)

	DescribeTable("rejects unsupported inputs",
		func(year, month, day int, region string) {
			_, err := calc.LunarToSolar(year, month, day, region)
			Expect(err).To(HaveOccurred())
		},
		Entry("unknown region", 2020, 1, 1, "xx"),
		Entry("year before the table starts", 1899, 1, 1, "cn"),
		Entry("year after the table ends", 2050, 1, 1, "cn"),
		Entry("month too low", 2020, 0, 1, "cn"),
		Entry("month too high", 2020, 13, 1, "cn"),
	)

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
