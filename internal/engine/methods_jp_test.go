package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/holidays/go-holidays/internal/definition"
)

var _ = Describe("jp methods", func() {
	call := func(name string, args MethodArgs) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(args)
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("jp_vernal_equinox_day falls in March", func() {
		got := call("jp_vernal_equinox_day", MethodArgs{Year: 2020})
		Expect(got.Month()).To(Equal(time.March))
	})

	It("jp_national_culture_day falls in September", func() {
		got := call("jp_national_culture_day", MethodArgs{Year: 2020})
		Expect(got.Month()).To(Equal(time.September))
	})

	It("jp_mountain_holiday is always August 11", func() {
		Expect(call("jp_mountain_holiday", MethodArgs{Year: 2020})).
			To(Equal(time.Date(2020, 8, 11, 0, 0, 0, 0, time.UTC)))
	})

	Describe("jp_citizens_holiday", func() {
		It("returns the day before the autumn equinox when it is a Wednesday", func() {
			Expect(call("jp_citizens_holiday", MethodArgs{Year: 2015})).
				To(Equal(time.Date(2015, 9, 22, 0, 0, 0, 0, time.UTC)))
		})

		It("returns zero when the autumn equinox is not a Wednesday", func() {
			Expect(call("jp_citizens_holiday", MethodArgs{Year: 2016}).IsZero()).To(BeTrue())
		})

		It("propagates a jpEquinox error", func() {
			fn, ok := LookupMethod("jp_citizens_holiday")
			Expect(ok).To(BeTrue())
			_, err := fn(MethodArgs{Year: 1700})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("jp_substitute_holiday", func() {
		It("returns zero when the date is zero", func() {
			Expect(call("jp_substitute_holiday", MethodArgs{Date: time.Time{}, Year: 2020}).IsZero()).To(BeTrue())
		})

		It("returns zero when the date is not a Sunday", func() {
			d := time.Date(2020, 8, 11, 0, 0, 0, 0, time.UTC) // Tuesday
			Expect(call("jp_substitute_holiday", MethodArgs{Date: d, Year: 2020}).IsZero()).To(BeTrue())
		})

		It("advances past a Sunday to the next non-Sunday, non-fixed-holiday day", func() {
			d := time.Date(2019, 8, 11, 0, 0, 0, 0, time.UTC) // Sunday
			got := call("jp_substitute_holiday", MethodArgs{Date: d, Year: 2019})
			Expect(got).To(Equal(time.Date(2019, 8, 12, 0, 0, 0, 0, time.UTC)))
		})
	})

	Describe("substitute wrappers", func() {
		It("jp_marine_day_substitute returns zero: the 3rd Monday of July is never a Sunday", func() {
			Expect(call("jp_marine_day_substitute", MethodArgs{Year: 2020}).IsZero()).To(BeTrue())
		})

		It("jp_health_sports_day_substitute returns zero: the 2nd Monday of October is never a Sunday", func() {
			Expect(call("jp_health_sports_day_substitute", MethodArgs{Year: 2020}).IsZero()).To(BeTrue())
		})

		It("jp_respect_for_aged_holiday_substitute returns zero: the 3rd Monday of September is never a Sunday", func() {
			Expect(call("jp_respect_for_aged_holiday_substitute", MethodArgs{Year: 2020}).IsZero()).To(BeTrue())
		})

		It("jp_mountain_holiday_substitute advances when August 11 falls on a Sunday", func() {
			// August 11, 2019 is a Sunday.
			got := call("jp_mountain_holiday_substitute", MethodArgs{Year: 2019})
			Expect(got).To(Equal(time.Date(2019, 8, 12, 0, 0, 0, 0, time.UTC)))
		})

		It("jp_mountain_holiday_substitute returns zero when August 11 is not a Sunday", func() {
			Expect(call("jp_mountain_holiday_substitute", MethodArgs{Year: 2020}).IsZero()).To(BeTrue())
		})

		It("jp_national_culture_day_substitute advances when the autumn equinox falls on a Sunday", func() {
			// The autumn equinox in 2018 falls on Sunday, September 23.
			got := call("jp_national_culture_day_substitute", MethodArgs{Year: 2018})
			Expect(got).To(Equal(time.Date(2018, 9, 24, 0, 0, 0, 0, time.UTC)))
		})

		It("jp_vernal_equinox_day_substitute returns zero when jpEquinox errors (base is zero)", func() {
			Expect(call("jp_vernal_equinox_day_substitute", MethodArgs{Year: 1700}).IsZero()).To(BeTrue())
		})
	})
})

var _ = Describe("jpEquinox", func() {
	It("errors for a year outside every supported base range", func() {
		_, err := jpEquinox(1700, time.March, vernalEquinoxBase)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("out of supported range"))
	})

	It("computes a date within each defined base range", func() {
		for _, year := range []int{1860, 1920, 2000, 2120} {
			got, err := jpEquinox(year, time.March, vernalEquinoxBase)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Year()).To(Equal(year))
		}
	})
})

var _ = Describe("jpFixedHolidaysByMonth", func() {
	AfterEach(func() {
		UnregisterCountry("jp")
	})

	It("collects only fixed mday rules that apply in the given year, keyed by month/day", func() {
		RegisterCountry("jp", []definition.HolidayRule{
			{Name: "Fixed Jan Day", Regions: []string{"jp"}, Month: 1, Mday: 15},
			{Name: "Another Fixed Jan Day", Regions: []string{"jp"}, Month: 1, Mday: 20},
			{Name: "Function Based", Regions: []string{"jp"}, Month: 3, Function: "jp_vernal_equinox_day"},
			{Name: "No Mday", Regions: []string{"jp"}, Month: 5, Wday: 1, Week: 1},
			{Name: "Year Restricted", Regions: []string{"jp"}, Month: 1, Mday: 25,
				YearRanges: []definition.YearRange{{Kind: definition.YearRangeUntil, Years: []int{2000}}}},
		})

		got := jpFixedHolidaysByMonth(2020)
		Expect(got[1][15]).To(BeTrue())
		Expect(got[1][20]).To(BeTrue())
		Expect(got[3]).To(BeNil())
		Expect(got[5]).To(BeNil())
		Expect(got[1][25]).To(BeFalse())
	})
})

var _ = Describe("jpNextWeekday", func() {
	AfterEach(func() {
		UnregisterCountry("jp")
	})

	It("skips Sundays and any registered fixed jp holiday", func() {
		RegisterCountry("jp", []definition.HolidayRule{
			{Name: "Fixed Holiday", Regions: []string{"jp"}, Month: 8, Mday: 12},
		})
		// Start on a Sunday; the next day (Aug 12) is itself a fixed holiday, so
		// the walk must continue past it to Aug 13.
		got := jpNextWeekday(time.Date(2019, 8, 11, 0, 0, 0, 0, time.UTC), 2019)
		Expect(got).To(Equal(time.Date(2019, 8, 13, 0, 0, 0, 0, time.UTC)))
	})

	It("returns the starting date immediately when it is neither a Sunday nor a fixed holiday", func() {
		d := time.Date(2020, 8, 12, 0, 0, 0, 0, time.UTC) // Wednesday
		Expect(jpNextWeekday(d, 2020)).To(Equal(d))
	})
})
