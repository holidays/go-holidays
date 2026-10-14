package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cl methods", func() {
	call := func(name string, args MethodArgs) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(args)
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Describe("st_peter_st_paul_cl", func() {
		It("rolls a Tue/Wed/Thu June 29 back to the preceding Monday", func() {
			// June 29, 2021 is a Tuesday.
			Expect(call("st_peter_st_paul_cl", MethodArgs{Year: 2021})).
				To(Equal(time.Date(2021, 6, 28, 0, 0, 0, 0, time.UTC)))
		})

		It("rolls a Friday June 29 forward to the following Monday", func() {
			// June 29, 2018 is a Friday.
			Expect(call("st_peter_st_paul_cl", MethodArgs{Year: 2018})).
				To(Equal(time.Date(2018, 7, 2, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a weekend/Monday June 29 unchanged", func() {
			// June 29, 2019 is a Saturday.
			Expect(call("st_peter_st_paul_cl", MethodArgs{Year: 2019})).
				To(Equal(time.Date(2019, 6, 29, 0, 0, 0, 0, time.UTC)))
		})
	})

	Describe("columbus_day_cl", func() {
		It("leaves an already-Monday October 12 unchanged", func() {
			// October 12, 2020 is a Monday.
			Expect(call("columbus_day_cl", MethodArgs{Year: 2020})).
				To(Equal(time.Date(2020, 10, 12, 0, 0, 0, 0, time.UTC)))
		})
	})

	Describe("other_churches_day_cl", func() {
		It("rolls a Tuesday October 31 back four days", func() {
			// October 31, 2017 is a Tuesday.
			Expect(call("other_churches_day_cl", MethodArgs{Year: 2017})).
				To(Equal(time.Date(2017, 10, 27, 0, 0, 0, 0, time.UTC)))
		})

		It("rolls a Wednesday October 31 forward two days", func() {
			// October 31, 2018 is a Wednesday.
			Expect(call("other_churches_day_cl", MethodArgs{Year: 2018})).
				To(Equal(time.Date(2018, 11, 2, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves any other weekday unchanged", func() {
			// October 31, 2016 is a Monday.
			Expect(call("other_churches_day_cl", MethodArgs{Year: 2016})).
				To(Equal(time.Date(2016, 10, 31, 0, 0, 0, 0, time.UTC)))
		})
	})

	Describe("indigenous_peoples_day_cl", func() {
		It("uses the statutory date in 2021", func() {
			Expect(call("indigenous_peoples_day_cl", MethodArgs{Year: 2021})).
				To(Equal(time.Date(2021, 6, 21, 0, 0, 0, 0, time.UTC)))
		})

		DescribeTable("matches the winter solstice in Chile",
			func(year, day int) {
				Expect(call("indigenous_peoples_day_cl", MethodArgs{Year: year})).
					To(Equal(time.Date(year, 6, day, 0, 0, 0, 0, time.UTC)))
			},
			Entry("2022", 2022, 21),
			Entry("2024", 2024, 20),
			Entry("2027", 2027, 21),
			Entry("2028", 2028, 20),
			// The 2029 solstice is 2029-06-21 01:48 UTC, still June 20 in Chile.
			Entry("2029 (Chile time, not UTC)", 2029, 20),
		)

		It("errors outside the supported range", func() {
			fn, ok := LookupMethod("indigenous_peoples_day_cl")
			Expect(ok).To(BeTrue())
			for _, year := range []int{2020, 2100} {
				_, err := fn(MethodArgs{Year: year})
				Expect(err).To(HaveOccurred(), "year %d", year)
			}
		})
	})
})
