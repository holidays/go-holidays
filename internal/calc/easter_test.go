package calc_test

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Easter", func() {
	DescribeTable("computes Gregorian Easter Sunday",
		func(year, month, day int) {
			got := calc.Easter(year)
			Expect(got).To(Equal(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)))
		},
		// Cross-checked against internal/definitions generated table tests
		// (gb, au, ca, us), which use the plain "easter" function and are
		// themselves oracle-verified.
		Entry("2008", 2008, 3, 23),
		Entry("2010", 2010, 4, 4),
		Entry("2013", 2013, 3, 31),
		Entry("2017", 2017, 4, 16),
		Entry("2018", 2018, 4, 1),
		Entry("2019", 2019, 4, 21),
	)
})

var _ = Describe("OrthodoxEaster", func() {
	DescribeTable("computes Orthodox Easter Sunday in the Gregorian calendar",
		func(year, month, day int) {
			got := calc.OrthodoxEaster(year)
			Expect(got).To(Equal(time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)))
		},
		// Cross-checked against internal/definitions generated table tests:
		// cy (orthodox_easter, "Κυριακή του Πάσχα") and bg_en (orthodox_easter,
		// "Easter Sunday" for the bg_en region), both oracle-verified.
		Entry("1985", 1985, 4, 14),
		Entry("2011", 2011, 4, 24),
		Entry("2015", 2015, 4, 12),
	)

	It("equals the Julian Easter date shifted by the century offset", func() {
		for _, year := range []int{1985, 2000, 2011, 2024} {
			julian := calc.OrthodoxEasterJulian(year)
			century := year / 100
			offset := century - century/4 - 2
			Expect(calc.OrthodoxEaster(year)).To(Equal(julian.AddDate(0, 0, offset)))
		}
	})
})

var _ = Describe("OrthodoxEasterJulian", func() {
	It("returns a date in late March or April, ahead of orthodox_easter's Julian->Gregorian shift", func() {
		got := calc.OrthodoxEasterJulian(2024)
		Expect(got.Year()).To(Equal(2024))
		Expect(got.Month()).To(BeNumerically(">=", time.March))
		Expect(got.Month()).To(BeNumerically("<=", time.April))
	})
})
