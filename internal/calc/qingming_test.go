package calc_test

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Qingming", func() {
	DescribeTable("computes the Gregorian date of the Qingming solar term",
		func(year, wantDay int) {
			got, err := calc.Qingming(year)
			Expect(err).NotTo(HaveOccurred())
			Expect(got).To(Equal(time.Date(year, time.April, wantDay, 0, 0, 0, 0, time.UTC)))
		},
		Entry("1999", 1999, 5),
		Entry("2008", 2008, 4),
		Entry("2009", 2009, 4),
		Entry("2010", 2010, 5),
		Entry("2011", 2011, 5),
		Entry("2012", 2012, 4),
		Entry("2013", 2013, 4),
		Entry("2014", 2014, 5),
		Entry("2020", 2020, 4),
		Entry("2021", 2021, 4),
		Entry("2022", 2022, 5),
		Entry("2023", 2023, 5),
		Entry("2024", 2024, 4),
		Entry("2025", 2025, 4),
		Entry("2026", 2026, 5),
	)

	DescribeTable("rejects years outside the supported range",
		func(year int) {
			_, err := calc.Qingming(year)
			Expect(err).To(HaveOccurred())
		},
		Entry("before 1900", 1899),
		Entry("after 2099", 2100),
	)
})
