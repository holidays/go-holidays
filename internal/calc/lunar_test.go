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
		Entry("vn Giỗ tổ Hùng Vương 2018 (Vietnamese table)", 2018, 3, 10, "vn", 2018, 3, 27),
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
})
