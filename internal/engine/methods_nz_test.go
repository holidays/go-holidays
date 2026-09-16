package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("nz methods", func() {
	call := func(name string, args MethodArgs) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(args)
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Describe("closest_monday", func() {
		It("rolls a Sunday forward one day", func() {
			d := time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC) // Sunday
			Expect(call("closest_monday", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)))
		})

		It("rolls Mon-Thu backward to the prior Monday", func() {
			d := time.Date(2020, 3, 5, 0, 0, 0, 0, time.UTC) // Thursday
			Expect(call("closest_monday", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)))
		})

		It("rolls Fri/Sat forward to the next Monday", func() {
			d := time.Date(2020, 3, 6, 0, 0, 0, 0, time.UTC) // Friday
			Expect(call("closest_monday", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)))
		})
	})

	It("previous_friday shifts back three days", func() {
		d := time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)
		Expect(call("previous_friday", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 6, 0, 0, 0, 0, time.UTC)))
	})

	It("next_week shifts forward seven days", func() {
		d := time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)
		Expect(call("next_week", MethodArgs{Date: d})).To(Equal(time.Date(2020, 3, 16, 0, 0, 0, 0, time.UTC)))
	})

	It("nz_canterbury_anniversary is the Friday after the second Tuesday of November", func() {
		got := call("nz_canterbury_anniversary", MethodArgs{Year: 2020})
		Expect(got.Weekday()).To(Equal(time.Friday))
		Expect(got.Month()).To(Equal(time.November))
	})

	Describe("matariki", func() {
		It("returns the legislated date for a known year", func() {
			Expect(call("matariki", MethodArgs{Year: 2023})).To(Equal(time.Date(2023, time.July, 14, 0, 0, 0, 0, time.UTC)))
		})

		It("returns zero for a year with no legislated date", func() {
			Expect(call("matariki", MethodArgs{Year: 1999}).IsZero()).To(BeTrue())
		})
	})
})
