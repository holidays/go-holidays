package calc_test

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DayOfMonth", func() {
	DescribeTable("returns the Nth occurrence of a weekday in a month",
		func(year int, month time.Month, week int, wday time.Weekday, wantMonth time.Month, wantDay int) {
			got := calc.DayOfMonth(year, month, week, wday)
			Expect(got).To(Equal(time.Date(year, wantMonth, wantDay, 0, 0, 0, 0, time.UTC)))
		},
		// US Thanksgiving: 4th Thursday of November.
		Entry("4th Thursday of Nov 2023", 2023, time.November, 4, time.Thursday, time.November, 23),
		// MLK Day: 3rd Monday of January.
		Entry("3rd Monday of Jan 2024", 2024, time.January, 3, time.Monday, time.January, 15),
		// 1st occurrence, where the month itself starts on the target weekday.
		Entry("1st Sunday of Sept 2024 (month starts on target weekday)", 2024, time.September, 1, time.Sunday, time.September, 1),
		// 5th occurrence that just barely exists (5 Fridays in the month).
		Entry("5th Friday of Aug 2025", 2025, time.August, 5, time.Friday, time.August, 29),
		// Last occurrence (week == -1).
		Entry("last Friday of Sept 2024", 2024, time.September, -1, time.Friday, time.September, 27),
	)

	It("returns the zero time when the Nth occurrence does not exist in the month", func() {
		// Feb 2023 has only 4 Mondays (6, 13, 20, 27); there is no 5th.
		got := calc.DayOfMonth(2023, time.February, 5, time.Monday)
		Expect(got).To(Equal(time.Time{}))
	})
})
