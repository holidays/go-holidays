package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Coverage for the ~15 well-known builtins registered in builtins.go's
// init(). Each closure is a thin wrapper around internal/calc; these tests
// just exercise every registered closure body once through LookupMethod,
// the same path ResolveYear uses at runtime.
var _ = Describe("builtins", func() {
	call := func(name string, args MethodArgs) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(args)
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("easter", func() {
		Expect(call("easter", MethodArgs{Year: 2020})).To(Equal(time.Date(2020, 4, 12, 0, 0, 0, 0, time.UTC)))
	})

	It("orthodox_easter", func() {
		Expect(call("orthodox_easter", MethodArgs{Year: 2020}).IsZero()).To(BeFalse())
	})

	It("orthodox_easter_julian", func() {
		Expect(call("orthodox_easter_julian", MethodArgs{Year: 2020}).IsZero()).To(BeFalse())
	})

	It("to_monday_if_sunday", func() {
		d := time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC) // Sunday
		Expect(call("to_monday_if_sunday", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)))
	})

	It("to_monday_if_weekend", func() {
		d := time.Date(2020, 3, 7, 0, 0, 0, 0, time.UTC) // Saturday
		Expect(call("to_monday_if_weekend", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)))
	})

	It("to_weekday_if_weekend", func() {
		d := time.Date(2020, 3, 7, 0, 0, 0, 0, time.UTC) // Saturday
		Expect(call("to_weekday_if_weekend", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 6, 0, 0, 0, 0, time.UTC)))
	})

	It("to_weekday_if_boxing_weekend", func() {
		d := time.Date(2025, 12, 27, 0, 0, 0, 0, time.UTC) // Saturday
		Expect(call("to_weekday_if_boxing_weekend", MethodArgs{Date: d}).IsZero()).To(BeFalse())
	})

	It("to_weekday_if_boxing_weekend_from_year", func() {
		// 2026-12-26 is a Saturday.
		Expect(call("to_weekday_if_boxing_weekend_from_year", MethodArgs{Year: 2026}).IsZero()).To(BeFalse())
	})

	It("to_weekday_if_boxing_weekend_from_year_or_to_tuesday_if_monday", func() {
		Expect(call("to_weekday_if_boxing_weekend_from_year_or_to_tuesday_if_monday", MethodArgs{Year: 2026}).IsZero()).To(BeFalse())
	})

	It("to_tuesday_if_sunday_or_monday_if_saturday", func() {
		d := time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC) // Sunday
		Expect(call("to_tuesday_if_sunday_or_monday_if_saturday", MethodArgs{Date: d}).IsZero()).To(BeFalse())
	})

	It("to_the_weekday_after", func() {
		d := time.Date(2020, 3, 6, 0, 0, 0, 0, time.UTC) // Friday
		Expect(call("to_the_weekday_after", MethodArgs{Date: d}).IsZero()).To(BeFalse())
	})

	It("to_the_second_weekday_after", func() {
		d := time.Date(2020, 3, 6, 0, 0, 0, 0, time.UTC)
		Expect(call("to_the_second_weekday_after", MethodArgs{Date: d}).IsZero()).To(BeFalse())
	})

	It("to_previous_day_if_leap_year", func() {
		d := time.Date(2020, 2, 29, 0, 0, 0, 0, time.UTC)
		Expect(call("to_previous_day_if_leap_year", MethodArgs{Date: d}).IsZero()).To(BeFalse())
	})

	It("lunar_to_solar", func() {
		got := call("lunar_to_solar", MethodArgs{Year: 2020, Month: 1, Day: 1, Region: "kr"})
		Expect(got.IsZero()).To(BeFalse())
	})

	It("lunar_to_solar propagates an error for an unsupported region", func() {
		fn, ok := LookupMethod("lunar_to_solar")
		Expect(ok).To(BeTrue())
		_, err := fn(MethodArgs{Year: 2020, Month: 1, Day: 1, Region: "zz_not_a_lunar_region"})
		Expect(err).To(HaveOccurred())
	})
})
