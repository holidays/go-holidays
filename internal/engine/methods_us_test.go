package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Regression coverage for independence_day: definitions master (commit
// 4488c2f) added this as the observed-shift function for US Independence
// Day, replacing the old us_va-only to_weekday_if_weekend hack. Behavior
// varies by region: us_ri shifts forward to the next Monday on a weekend,
// us_tx never shifts, and everyone else uses nearest-weekday.
var _ = Describe("independence_day", func() {
	var fn Method

	BeforeEach(func() {
		var ok bool
		fn, ok = LookupMethod("independence_day")
		Expect(ok).To(BeTrue(), "independence_day must be registered")
	})

	call := func(region string, date time.Time) time.Time {
		got, err := fn(MethodArgs{Region: region, Date: date})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Context("default regions (us, us_va, ...)", func() {
		It("shifts a Saturday back to Friday", func() {
			Expect(call("us", time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC))).
				To(Equal(time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)))
		})

		It("shifts a Sunday forward to Monday", func() {
			Expect(call("us_va", time.Date(2027, 7, 4, 0, 0, 0, 0, time.UTC))).
				To(Equal(time.Date(2027, 7, 5, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a weekday unchanged", func() {
			d := time.Date(2028, 7, 4, 0, 0, 0, 0, time.UTC) // Tuesday
			Expect(call("us", d)).To(Equal(d))
		})
	})

	Context("us_ri", func() {
		It("shifts a Saturday forward two days to Monday", func() {
			Expect(call("us_ri", time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC))).
				To(Equal(time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)))
		})

		It("shifts a Sunday forward one day to Monday", func() {
			Expect(call("us_ri", time.Date(2027, 7, 4, 0, 0, 0, 0, time.UTC))).
				To(Equal(time.Date(2027, 7, 5, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a weekday unchanged", func() {
			d := time.Date(2028, 7, 4, 0, 0, 0, 0, time.UTC)
			Expect(call("us_ri", d)).To(Equal(d))
		})
	})

	Context("us_tx", func() {
		It("never shifts, even on a weekend", func() {
			d := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC) // Saturday
			Expect(call("us_tx", d)).To(Equal(d))
		})
	})
})

var _ = Describe("christmas_eve_holiday", func() {
	call := func(d time.Time) time.Time {
		fn, ok := LookupMethod("christmas_eve_holiday")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Date: d})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("shifts a Saturday back one day", func() {
		d := time.Date(2022, 12, 24, 0, 0, 0, 0, time.UTC) // Saturday
		Expect(call(d)).To(Equal(time.Date(2022, 12, 23, 0, 0, 0, 0, time.UTC)))
	})

	It("shifts a Sunday back two days", func() {
		d := time.Date(2023, 12, 24, 0, 0, 0, 0, time.UTC) // Sunday
		Expect(call(d)).To(Equal(time.Date(2023, 12, 22, 0, 0, 0, 0, time.UTC)))
	})

	It("leaves a weekday unchanged", func() {
		d := time.Date(2021, 12, 24, 0, 0, 0, 0, time.UTC) // Friday
		Expect(call(d)).To(Equal(d))
	})
})

var _ = DescribeTable("Jewish-calendar holiday tables",
	func(name string, year int, want time.Time) {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(want))
	},
	Entry("rosh_hashanah known year", "rosh_hashanah", 2016, time.Date(2016, 10, 3, 0, 0, 0, 0, time.UTC)),
	Entry("yom_kippur known year", "yom_kippur", 2016, time.Date(2016, 10, 12, 0, 0, 0, 0, time.UTC)),
)

var _ = Describe("Jewish-calendar holiday tables outside their range", func() {
	It("rosh_hashanah returns zero for a year outside the table", func() {
		fn, ok := LookupMethod("rosh_hashanah")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 1999})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.IsZero()).To(BeTrue())
	})

	It("yom_kippur returns zero for a year outside the table", func() {
		fn, ok := LookupMethod("yom_kippur")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 1999})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.IsZero()).To(BeTrue())
	})
})

var _ = Describe("georgia_state_holiday", func() {
	It("rolls back to the Monday before April 26 when April 26 is not itself a Monday", func() {
		// April 26, 2022 is a Tuesday.
		fn, ok := LookupMethod("georgia_state_holiday")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2022, Month: 4})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2022, 4, 25, 0, 0, 0, 0, time.UTC)))
	})

	It("returns April 26 itself when it already falls on a Monday", func() {
		fn, ok := LookupMethod("georgia_state_holiday")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2021, Month: 4})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2021, 4, 26, 0, 0, 0, 0, time.UTC)))
	})
})

var _ = Describe("lee_jackson_day", func() {
	It("is the Friday before the third Monday in January", func() {
		fn, ok := LookupMethod("lee_jackson_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2021, Month: 1})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Weekday()).To(Equal(time.Friday))
	})
})

var _ = Describe("election_day", func() {
	It("is the Tuesday after the first Monday of November", func() {
		fn, ok := LookupMethod("election_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2020})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Weekday()).To(Equal(time.Tuesday))
	})
})

var _ = Describe("even_year_election_day", func() {
	call := func(year int) time.Time {
		fn, ok := LookupMethod("even_year_election_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("returns zero for an odd year", func() {
		Expect(call(2021).IsZero()).To(BeTrue())
	})

	It("returns the election-day date for an even year", func() {
		got := call(2020)
		Expect(got.Weekday()).To(Equal(time.Tuesday))
	})
})

var _ = Describe("us_inauguration_day", func() {
	call := func(year int) time.Time {
		fn, ok := LookupMethod("us_inauguration_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("returns zero for a non-inauguration year", func() {
		Expect(call(2020).IsZero()).To(BeTrue())
	})

	It("returns January 20 in the year following a presidential election", func() {
		Expect(call(2021)).To(Equal(time.Date(2021, 1, 20, 0, 0, 0, 0, time.UTC)))
	})
})

var _ = Describe("day_after_thanksgiving", func() {
	It("is the day after the 4th Thursday of November", func() {
		fn, ok := LookupMethod("day_after_thanksgiving")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2020})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Weekday()).To(Equal(time.Friday))
	})
})

var _ = Describe("juneteenth_national_independence_day", func() {
	call := func(region string, d time.Time) time.Time {
		fn, ok := LookupMethod("juneteenth_national_independence_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Region: region, Date: d})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Context("us_ut", func() {
		It("shifts a Sunday forward one day", func() {
			d := time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC) // Sunday
			Expect(call("us_ut", d)).To(Equal(time.Date(2023, 6, 19, 0, 0, 0, 0, time.UTC)))
		})

		It("shifts a Saturday forward two days", func() {
			d := time.Date(2022, 6, 18, 0, 0, 0, 0, time.UTC) // Saturday
			Expect(call("us_ut", d)).To(Equal(time.Date(2022, 6, 20, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a Monday unchanged", func() {
			d := time.Date(2029, 6, 18, 0, 0, 0, 0, time.UTC) // Monday
			Expect(call("us_ut", d)).To(Equal(d))
		})

		It("rolls any other weekday back to the preceding Monday", func() {
			d := time.Date(2024, 6, 19, 0, 0, 0, 0, time.UTC) // Wednesday
			Expect(call("us_ut", d)).To(Equal(time.Date(2024, 6, 17, 0, 0, 0, 0, time.UTC)))
		})
	})

	Context("default regions", func() {
		It("shifts a Sunday forward one day", func() {
			d := time.Date(2023, 6, 18, 0, 0, 0, 0, time.UTC) // Sunday
			Expect(call("us", d)).To(Equal(time.Date(2023, 6, 19, 0, 0, 0, 0, time.UTC)))
		})

		It("shifts a Saturday back one day", func() {
			d := time.Date(2022, 6, 18, 0, 0, 0, 0, time.UTC) // Saturday
			Expect(call("us", d)).To(Equal(time.Date(2022, 6, 17, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a weekday unchanged", func() {
			d := time.Date(2024, 6, 19, 0, 0, 0, 0, time.UTC) // Wednesday
			Expect(call("us", d)).To(Equal(d))
		})
	})
})
