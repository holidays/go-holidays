package calc_test

import (
	"errors"
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
		// Years where Vietnam (UTC+7) and Korea (UTC+9) disagree, which the
		// old copy of the Korean table got wrong.
		Entry("vn Tết 1985 (Vietnamese table, UTC+7)", 1985, 1, 1, "vn", 1985, 1, 21),
		Entry("vn Tết 2007 (Vietnamese table, UTC+7)", 2007, 1, 1, "vn", 2007, 2, 17),
		Entry("vn Giỗ tổ Hùng Vương 2009 (Vietnamese table, UTC+7)", 2009, 3, 10, "vn", 2009, 4, 4),
		Entry("kr lunar month 11 of 2042 (KASI new moon 23:29 KST)", 2042, 11, 1, "kr", 2042, 12, 12),
		Entry("kr Seollal 2043 (unaffected by the 2042 row fix)", 2043, 1, 1, "kr", 2043, 2, 10),
		// Rows after 2049. Expected dates come from sources independent of
		// the DE440 rows: the Hong Kong Observatory tables (cn, hk, sg), ICU's
		// Korean calendar (kr) and Ho Ngoc Duc's algorithm at UTC+7 (vn).
		Entry("cn Chinese New Year 2050 (HKO)", 2050, 1, 1, "cn", 2050, 1, 23),
		Entry("hk Chung Yeung 2057 (HKO, new moon near midnight)", 2057, 9, 9, "hk", 2057, 10, 6),
		Entry("cn Mid-Autumn 2075 (HKO)", 2075, 8, 15, "cn", 2075, 9, 24),
		Entry("sg Dragon Boat 2099 (HKO)", 2099, 5, 5, "sg", 2099, 6, 23),
		Entry("cn Mid-Autumn 2100 (HKO)", 2100, 8, 15, "cn", 2100, 9, 18),
		Entry("kr Seollal 2050 (ICU)", 2050, 1, 1, "kr", 2050, 1, 23),
		Entry("kr Chuseok 2075 (ICU)", 2075, 8, 15, "kr", 2075, 9, 24),
		Entry("kr Buddha's Birthday 2100 (ICU)", 2100, 4, 8, "kr", 2100, 5, 16),
		Entry("vn Tết 2072", 2072, 1, 1, "vn", 2072, 2, 19),
		Entry("vn Giỗ tổ Hùng Vương 2085", 2085, 3, 10, "vn", 2085, 4, 3),
		Entry("vn Giỗ tổ Hùng Vương 2100", 2100, 3, 10, "vn", 2100, 4, 18),
		// First and last rows of the lookup tables (1900 and 2100).
		Entry("earliest supported year (yearDiff == 0)", 1900, 1, 1, "cn", 1900, 1, 31),
		Entry("latest supported year (yearDiff == len-1)", 2100, 1, 1, "cn", 2100, 2, 9),
		Entry("last day of the latest supported year", 2100, 12, 29, "cn", 2101, 1, 28),
	)

	DescribeTable("returns ErrUnsupportedYear outside the tables",
		func(year int, region, wantMessage string) {
			_, err := calc.LunarToSolar(year, 1, 1, region)
			Expect(errors.Is(err, calc.ErrUnsupportedYear)).To(BeTrue())
			Expect(err.Error()).To(Equal(wantMessage))
		},
		Entry("year after the table ends", 2101, "kr",
			`calc.LunarToSolar: lunar calendar data for "kr" covers lunar years 1900..2100, got 2101: year outside supported range`),
		Entry("year before the table starts", 1899, "cn",
			`calc.LunarToSolar: lunar calendar data for "cn" covers lunar years 1900..2100, got 1899: year outside supported range`),
	)

	DescribeTable("rejects unsupported inputs",
		func(year, month, day int, region string) {
			_, err := calc.LunarToSolar(year, month, day, region)
			Expect(err).To(HaveOccurred())
		},
		Entry("unknown region", 2020, 1, 1, "xx"),
		Entry("year before the table starts", 1899, 1, 1, "cn"),
		Entry("year after the table ends", 2101, 1, 1, "cn"),
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
