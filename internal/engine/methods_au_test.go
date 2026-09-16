package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("au methods", func() {
	call := func(name string, args MethodArgs) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(args)
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Describe("afl_grand_final", func() {
		It("returns the hard-coded date for a known year", func() {
			Expect(call("afl_grand_final", MethodArgs{Year: 2015})).To(Equal(time.Date(2015, 10, 2, 0, 0, 0, 0, time.UTC)))
			Expect(call("afl_grand_final", MethodArgs{Year: 2016})).To(Equal(time.Date(2016, 9, 30, 0, 0, 0, 0, time.UTC)))
			Expect(call("afl_grand_final", MethodArgs{Year: 2017})).To(Equal(time.Date(2017, 9, 29, 0, 0, 0, 0, time.UTC)))
			Expect(call("afl_grand_final", MethodArgs{Year: 2020})).To(Equal(time.Date(2020, 10, 23, 0, 0, 0, 0, time.UTC)))
			Expect(call("afl_grand_final", MethodArgs{Year: 2022})).To(Equal(time.Date(2022, 9, 23, 0, 0, 0, 0, time.UTC)))
		})

		It("falls back to the last Friday of September for an unknown year", func() {
			Expect(call("afl_grand_final", MethodArgs{Year: 2021})).To(Equal(time.Date(2021, 9, 24, 0, 0, 0, 0, time.UTC)))
		})
	})

	It("qld_queens_bday_october", func() {
		Expect(call("qld_queens_bday_october", MethodArgs{Year: 2020}).Weekday()).To(Equal(time.Monday))
	})

	It("qld_kings_bday_october", func() {
		Expect(call("qld_kings_bday_october", MethodArgs{Year: 2020}).Weekday()).To(Equal(time.Monday))
	})

	It("qld_queens_birthday_june", func() {
		Expect(call("qld_queens_birthday_june", MethodArgs{Year: 2020}).Weekday()).To(Equal(time.Monday))
	})

	It("qld_labour_day_may", func() {
		Expect(call("qld_labour_day_may", MethodArgs{Year: 2020}).Weekday()).To(Equal(time.Monday))
	})

	It("qld_labour_day_october", func() {
		Expect(call("qld_labour_day_october", MethodArgs{Year: 2020}).Weekday()).To(Equal(time.Monday))
	})

	It("hobart_show_day", func() {
		got := call("hobart_show_day", MethodArgs{Year: 2020})
		Expect(got.Weekday()).To(Equal(time.Thursday))
	})

	It("march_pub_hol_sa", func() {
		Expect(call("march_pub_hol_sa", MethodArgs{Year: 2020}).Weekday()).To(Equal(time.Monday))
	})

	It("may_pub_hol_sa", func() {
		Expect(call("may_pub_hol_sa", MethodArgs{Year: 2020}).Weekday()).To(Equal(time.Monday))
	})

	Describe("qld_brisbane_ekka_holiday", func() {
		It("uses the second Friday when the first Friday of August falls before the 5th", func() {
			// August 1, 2025 is a Friday (day 1 < 5).
			got := call("qld_brisbane_ekka_holiday", MethodArgs{Year: 2025})
			Expect(got).To(Equal(time.Date(2025, 8, 13, 0, 0, 0, 0, time.UTC)))
		})

		It("uses the first Friday when it falls on or after the 5th", func() {
			// August 1, 2020 is a Saturday, so the first Friday is August 7.
			got := call("qld_brisbane_ekka_holiday", MethodArgs{Year: 2020})
			Expect(got).To(Equal(time.Date(2020, 8, 12, 0, 0, 0, 0, time.UTC)))
		})
	})

	Describe("to_nearest_monday_after", func() {
		It("rolls a Sunday forward one day", func() {
			d := time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC) // Sunday
			Expect(call("to_nearest_monday_after", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a Monday unchanged", func() {
			d := time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC) // Monday
			Expect(call("to_nearest_monday_after", MethodArgs{Date: d})).To(Equal(d))
		})

		It("rolls a mid-week day forward to the next Monday", func() {
			d := time.Date(2020, 3, 4, 0, 0, 0, 0, time.UTC) // Wednesday
			Expect(call("to_nearest_monday_after", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)))
		})
	})
})
