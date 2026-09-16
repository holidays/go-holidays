package calc_test

import (
	"time"

	"github.com/holidays/go-holidays/internal/calc"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Reference week: 2024-06-03 (Monday) through 2024-06-09 (Sunday).
func d(day int) time.Time {
	return time.Date(2024, time.June, day, 0, 0, 0, 0, time.UTC)
}

var _ = Describe("ToMondayIfSunday", func() {
	It("rolls Sunday forward to Monday", func() {
		Expect(calc.ToMondayIfSunday(d(9))).To(Equal(d(10)))
	})
	It("leaves other weekdays unchanged", func() {
		Expect(calc.ToMondayIfSunday(d(5))).To(Equal(d(5)))
	})
})

var _ = Describe("ToMondayIfWeekend", func() {
	It("rolls Saturday forward to Monday", func() {
		Expect(calc.ToMondayIfWeekend(d(8))).To(Equal(d(10)))
	})
	It("rolls Sunday forward to Monday", func() {
		Expect(calc.ToMondayIfWeekend(d(9))).To(Equal(d(10)))
	})
	It("leaves weekdays unchanged", func() {
		Expect(calc.ToMondayIfWeekend(d(4))).To(Equal(d(4)))
	})
})

var _ = Describe("ToWeekdayIfWeekend", func() {
	It("rolls Saturday back to Friday", func() {
		Expect(calc.ToWeekdayIfWeekend(d(8))).To(Equal(d(7)))
	})
	It("rolls Sunday forward to Monday", func() {
		Expect(calc.ToWeekdayIfWeekend(d(9))).To(Equal(d(10)))
	})
	It("leaves weekdays unchanged", func() {
		Expect(calc.ToWeekdayIfWeekend(d(4))).To(Equal(d(4)))
	})
})

var _ = Describe("ToWeekdayIfBoxingWeekend", func() {
	It("rolls Saturday forward by 2 days to Monday", func() {
		Expect(calc.ToWeekdayIfBoxingWeekend(d(8))).To(Equal(d(10)))
	})
	It("rolls Sunday forward by 2 days to Tuesday", func() {
		Expect(calc.ToWeekdayIfBoxingWeekend(d(9))).To(Equal(d(11)))
	})
	It("rolls Monday forward by 1 day to Tuesday", func() {
		Expect(calc.ToWeekdayIfBoxingWeekend(d(3))).To(Equal(d(4)))
	})
	It("leaves other weekdays unchanged", func() {
		Expect(calc.ToWeekdayIfBoxingWeekend(d(5))).To(Equal(d(5)))
	})
})

var _ = Describe("ToTuesdayIfSundayOrMondayIfSaturday", func() {
	It("rolls Saturday forward by 2 days to Monday", func() {
		Expect(calc.ToTuesdayIfSundayOrMondayIfSaturday(d(8))).To(Equal(d(10)))
	})
	It("rolls Sunday forward by 2 days to Tuesday", func() {
		Expect(calc.ToTuesdayIfSundayOrMondayIfSaturday(d(9))).To(Equal(d(11)))
	})
	It("leaves other weekdays unchanged", func() {
		Expect(calc.ToTuesdayIfSundayOrMondayIfSaturday(d(6))).To(Equal(d(6)))
	})
})

var _ = Describe("ToTheWeekdayAfter", func() {
	It("returns the plain next day when neither date nor date+1 is Sunday", func() {
		Expect(calc.ToTheWeekdayAfter(d(3))).To(Equal(d(4))) // Mon -> Tue
	})
	It("skips a Sunday landed on after adding a day", func() {
		Expect(calc.ToTheWeekdayAfter(d(8))).To(Equal(d(10))) // Sat -> Sun -> Mon
	})
	It("skips date itself when it is a Sunday, then adds a day", func() {
		Expect(calc.ToTheWeekdayAfter(d(9))).To(Equal(d(11))) // Sun -> Mon -> Tue
	})
})

var _ = Describe("ToTheSecondWeekdayAfter", func() {
	It("returns the day after ToTheWeekdayAfter when that is not a Sunday", func() {
		Expect(calc.ToTheSecondWeekdayAfter(d(3))).To(Equal(d(5))) // Mon -> Tue -> Wed
	})
	It("skips a Sunday landed on by the extra day", func() {
		Expect(calc.ToTheSecondWeekdayAfter(d(7))).To(Equal(d(10))) // Fri -> Sat -> Sun -> Mon
	})
})

var _ = Describe("ToPreviousDayIfLeapYear", func() {
	It("shifts back one day in a leap year", func() {
		leapDate := time.Date(2024, time.March, 1, 0, 0, 0, 0, time.UTC)
		Expect(calc.ToPreviousDayIfLeapYear(leapDate)).To(Equal(time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC)))
	})
	It("leaves the date unchanged in a non-leap year", func() {
		nonLeapDate := time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)
		Expect(calc.ToPreviousDayIfLeapYear(nonLeapDate)).To(Equal(nonLeapDate))
	})
})

var _ = Describe("IsLeapYear", func() {
	DescribeTable("applies the standard Gregorian leap year rule",
		func(year int, want bool) {
			Expect(calc.IsLeapYear(year)).To(Equal(want))
		},
		Entry("not divisible by 4", 2023, false),
		Entry("divisible by 4, not by 100", 2024, true),
		Entry("divisible by 100, not by 400", 1900, false),
		Entry("divisible by 400", 2000, true),
	)
})
