package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("co methods", func() {
	call := func(name string, args MethodArgs) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(args)
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Describe("to_following_monday_if_not_monday", func() {
		It("rolls a Sunday forward one day", func() {
			d := time.Date(2020, 3, 1, 0, 0, 0, 0, time.UTC) // Sunday
			Expect(call("to_following_monday_if_not_monday", MethodArgs{Date: d})).
				To(Equal(time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a Monday unchanged", func() {
			d := time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC) // Monday
			Expect(call("to_following_monday_if_not_monday", MethodArgs{Date: d})).To(Equal(d))
		})

		It("rolls any other weekday forward to the following Monday", func() {
			d := time.Date(2020, 3, 4, 0, 0, 0, 0, time.UTC) // Wednesday
			Expect(call("to_following_monday_if_not_monday", MethodArgs{Date: d})).
				To(Equal(time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)))
		})
	})

	DescribeTable("named Catholic-holiday anchors always land on a Monday",
		func(name string, year int) {
			got := call(name, MethodArgs{Year: year})
			Expect(got.Weekday()).To(Equal(time.Monday))
		},
		Entry("epiphany", "epiphany", 2019),
		Entry("saint_josephs_day", "saint_josephs_day", 2020),
		Entry("saint_peter_and_saint_paul", "saint_peter_and_saint_paul", 2020),
		Entry("assumption_of_mary", "assumption_of_mary", 2020),
		Entry("columbus_day", "columbus_day", 2020),
		Entry("all_saints_day", "all_saints_day", 2020),
		Entry("independence_of_cartagena", "independence_of_cartagena", 2020),
	)

	It("epiphany rolls a Sunday anchor forward to Monday", func() {
		// January 6, 2019 is a Sunday.
		Expect(call("epiphany", MethodArgs{Year: 2019})).
			To(Equal(time.Date(2019, 1, 7, 0, 0, 0, 0, time.UTC)))
	})

	It("epiphany leaves an already-Monday anchor unchanged", func() {
		// January 6, 2020 is a Monday.
		Expect(call("epiphany", MethodArgs{Year: 2020})).
			To(Equal(time.Date(2020, 1, 6, 0, 0, 0, 0, time.UTC)))
	})

	It("epiphany rolls a mid-week anchor forward to the following Monday", func() {
		// January 6, 2021 is a Wednesday.
		Expect(call("epiphany", MethodArgs{Year: 2021})).
			To(Equal(time.Date(2021, 1, 11, 0, 0, 0, 0, time.UTC)))
	})
})
