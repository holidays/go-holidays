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
})
