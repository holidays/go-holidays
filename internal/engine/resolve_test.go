package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/holidays/go-holidays/internal/definition"
)

var _ = Describe("ResolveYear", func() {
	AfterEach(func() {
		UnregisterCountry("zzres")
	})

	It("skips a rule that does not apply in the requested year", func() {
		RegisterCountry("zzres", []definition.HolidayRule{
			{Name: "Future Only", Regions: []string{"zzres"}, Month: 1, Mday: 1,
				YearRanges: []definition.YearRange{{Kind: definition.YearRangeFrom, Years: []int{3000}}}},
		})
		got, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zzres"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	It("skips an informal rule unless Informal is requested", func() {
		RegisterCountry("zzres", []definition.HolidayRule{
			{Name: "Informal Day", Regions: []string{"zzres"}, Month: 1, Mday: 1, Type: definition.Informal},
		})
		got, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zzres"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())

		got, err = ResolveYear(2020, ResolveOptions{Regions: []string{"zzres"}, Informal: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(1))
		Expect(got[0].Informal).To(BeTrue())
	})

	It("skips a rule that does not match the requested regions", func() {
		RegisterCountry("zzres", []definition.HolidayRule{
			{Name: "Other Region", Regions: []string{"zzres_other"}, Month: 1, Mday: 1},
		})
		got, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zzres_here"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	It("propagates a computeDate error, wrapped with the rule name", func() {
		RegisterCountry("zzres", []definition.HolidayRule{
			{Name: "Broken Function", Regions: []string{"zzres"}, Month: 1, Function: "zz_does_not_exist_fn"},
		})
		_, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zzres"}})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Broken Function"))
	})

	It("skips a rule whose computed date is zero", func() {
		RegisterCountry("zzres", []definition.HolidayRule{
			// 5th Monday of a month that only has 4: computeDate returns zero.
			{Name: "No Fifth Monday", Regions: []string{"zzres"}, Month: 2, Wday: 1, Week: 5},
		})
		got, err := ResolveYear(2021, ResolveOptions{Regions: []string{"zzres"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeEmpty())
	})

	It("applies the observed shift when Observed is requested and the rule has one", func() {
		RegisterMethod("zz_shift_one_day", func(a MethodArgs) (time.Time, error) {
			return a.Date.AddDate(0, 0, 1), nil
		})
		RegisterCountry("zzres", []definition.HolidayRule{
			{Name: "Shifted", Regions: []string{"zzres"}, Month: 1, Mday: 1, Observed: "zz_shift_one_day"},
		})
		got, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zzres"}, Observed: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(1))
		Expect(got[0].Date).To(Equal(time.Date(2020, 1, 2, 0, 0, 0, 0, time.UTC)))
	})

	It("does not apply the observed shift when Observed is not requested", func() {
		RegisterMethod("zz_shift_two_days", func(a MethodArgs) (time.Time, error) {
			return a.Date.AddDate(0, 0, 2), nil
		})
		RegisterCountry("zzres", []definition.HolidayRule{
			{Name: "Not Shifted", Regions: []string{"zzres"}, Month: 1, Mday: 1, Observed: "zz_shift_two_days"},
		})
		got, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zzres"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(1))
		Expect(got[0].Date).To(Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)))
	})

	It("propagates an applyObserved error, wrapped with the rule name and method", func() {
		RegisterCountry("zzres", []definition.HolidayRule{
			{Name: "Bad Observed", Regions: []string{"zzres"}, Month: 1, Mday: 1, Observed: "zz_observed_missing"},
		})
		_, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zzres"}, Observed: true})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("Bad Observed"))
		Expect(err.Error()).To(ContainSubstring("zz_observed_missing"))
	})

	It("dedups identical definitions that differ only by region, merging nothing extra", func() {
		RegisterCountry("zzres", []definition.HolidayRule{
			{Name: "Shared Day", Regions: []string{"zzres_a"}, Month: 3, Mday: 3},
			{Name: "Shared Day", Regions: []string{"zzres_b"}, Month: 3, Mday: 3},
		})
		got, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zzres_a", "zzres_b"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(HaveLen(1))
	})
})

var _ = Describe("computeDate", func() {
	It("computes a fixed month/day date", func() {
		rule := definition.HolidayRule{Name: "Fixed", Month: 6, Mday: 15}
		got, err := computeDate(rule, 2020)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2020, 6, 15, 0, 0, 0, 0, time.UTC)))
	})

	It("computes an nth-weekday date", func() {
		rule := definition.HolidayRule{Name: "Nth Weekday", Month: 9, Wday: 1, Week: 1} // first Monday of September
		got, err := computeDate(rule, 2020)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2020, 9, 7, 0, 0, 0, 0, time.UTC)))
	})

	It("returns zero, no error, when the nth weekday does not exist in the month", func() {
		rule := definition.HolidayRule{Name: "Fifth Monday", Month: 2, Wday: 1, Week: 5}
		got, err := computeDate(rule, 2021) // Feb 2021 has only 4 Mondays
		Expect(err).NotTo(HaveOccurred())
		Expect(got.IsZero()).To(BeTrue())
	})

	It("errors when the rule's function is not registered", func() {
		rule := definition.HolidayRule{Name: "Unregistered Fn", Month: 1, Function: "zz_totally_unregistered"}
		_, err := computeDate(rule, 2020)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unregistered method"))
	})

	It("propagates an error returned by the rule's function", func() {
		RegisterMethod("zz_erroring_fn", func(a MethodArgs) (time.Time, error) {
			return time.Time{}, errFixture
		})
		rule := definition.HolidayRule{Name: "Erroring Fn", Month: 1, Function: "zz_erroring_fn"}
		_, err := computeDate(rule, 2020)
		Expect(err).To(Equal(errFixture))
	})

	It("passes the rule's first region to the function", func() {
		var seenRegion string
		RegisterMethod("zz_region_capture", func(a MethodArgs) (time.Time, error) {
			seenRegion = a.Region
			return time.Date(a.Year, 1, 1, 0, 0, 0, 0, time.UTC), nil
		})
		rule := definition.HolidayRule{Name: "Region Capture", Month: 1, Function: "zz_region_capture", Regions: []string{"zzres_capture", "zzres_other"}}
		_, err := computeDate(rule, 2020)
		Expect(err).NotTo(HaveOccurred())
		Expect(seenRegion).To(Equal("zzres_capture"))
	})

	It("applies a function modifier as a day offset after a fixed date", func() {
		rule := definition.HolidayRule{Name: "Modified Fixed", Month: 6, Mday: 1, FunctionModifier: 3}
		got, err := computeDate(rule, 2020)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2020, 6, 4, 0, 0, 0, 0, time.UTC)))
	})

	It("errors when the rule has no mday, wday, or function", func() {
		rule := definition.HolidayRule{Name: "Empty Rule", Month: 1}
		_, err := computeDate(rule, 2020)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("no mday, wday/week, or function"))
	})

	It("pulls a function result that lands in a different year back into the resolution year", func() {
		RegisterMethod("zz_next_year_date", func(a MethodArgs) (time.Time, error) {
			// e.g. a lunar-month-12 eve landing on Jan 1 of the following year.
			return time.Date(a.Year+1, 1, 1, 0, 0, 0, 0, time.UTC), nil
		})
		rule := definition.HolidayRule{Name: "Cross Year", Month: 12, Function: "zz_next_year_date"}
		got, err := computeDate(rule, 2020)
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)))
	})

	It("does not force the year for a zero result produced by a function", func() {
		RegisterMethod("zz_zero_result_fn", func(a MethodArgs) (time.Time, error) {
			return time.Time{}, nil
		})
		rule := definition.HolidayRule{Name: "Zero Fn", Month: 1, Function: "zz_zero_result_fn"}
		got, err := computeDate(rule, 2020)
		Expect(err).NotTo(HaveOccurred())
		Expect(got.IsZero()).To(BeTrue())
	})
})

var _ = Describe("requestedRegion", func() {
	It("returns empty string when no regions were requested", func() {
		Expect(requestedRegion(nil)).To(Equal(""))
	})

	It("returns the first requested region", func() {
		Expect(requestedRegion([]string{"us_ga", "us_ny"})).To(Equal("us_ga"))
	})
})

var _ = Describe("applyObserved", func() {
	It("errors when the observed method is not registered", func() {
		_, err := applyObserved("zz_observed_not_registered", time.Now(), "")
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unregistered observed method"))
	})

	It("invokes the registered method with the date broken into fields", func() {
		RegisterMethod("zz_observed_capture", func(a MethodArgs) (time.Time, error) {
			Expect(a.Year).To(Equal(2020))
			Expect(a.Month).To(Equal(7))
			Expect(a.Day).To(Equal(4))
			Expect(a.Region).To(Equal("us_ga"))
			return a.Date.AddDate(0, 0, 1), nil
		})
		got, err := applyObserved("zz_observed_capture", time.Date(2020, 7, 4, 0, 0, 0, 0, time.UTC), "us_ga")
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2020, 7, 5, 0, 0, 0, 0, time.UTC)))
	})
})

var _ = Describe("nthOrLastWeekday", func() {
	It("returns the last matching weekday of the month when week is -1", func() {
		got := nthOrLastWeekday(2020, time.September, -1, time.Wednesday)
		Expect(got).To(Equal(time.Date(2020, 9, 30, 0, 0, 0, 0, time.UTC)))
	})

	It("returns the nth matching weekday of the month", func() {
		got := nthOrLastWeekday(2020, time.September, 2, time.Wednesday)
		Expect(got).To(Equal(time.Date(2020, 9, 9, 0, 0, 0, 0, time.UTC)))
	})

	It("returns zero when the nth weekday overflows past the end of the month", func() {
		got := nthOrLastWeekday(2021, time.February, 5, time.Monday)
		Expect(got.IsZero()).To(BeTrue())
	})
})

var errFixture = fixtureError("fixture error")

type fixtureError string

func (e fixtureError) Error() string { return string(e) }
