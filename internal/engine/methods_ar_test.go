package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("to_nearest_monday (ar)", func() {
	var fn Method

	BeforeEach(func() {
		var ok bool
		fn, ok = LookupMethod("to_nearest_monday")
		Expect(ok).To(BeTrue())
	})

	call := func(fn Method, d time.Time) time.Time {
		got, err := fn(MethodArgs{Date: d})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("rolls Tuesday back one day", func() {
		d := time.Date(2020, 3, 3, 0, 0, 0, 0, time.UTC) // Tuesday
		Expect(call(fn, d)).To(Equal(time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)))
	})

	It("rolls Wednesday back two days", func() {
		d := time.Date(2020, 3, 4, 0, 0, 0, 0, time.UTC) // Wednesday
		Expect(call(fn, d)).To(Equal(time.Date(2020, 3, 2, 0, 0, 0, 0, time.UTC)))
	})

	It("rolls Thursday forward four days", func() {
		d := time.Date(2020, 3, 5, 0, 0, 0, 0, time.UTC) // Thursday
		Expect(call(fn, d)).To(Equal(time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)))
	})

	It("rolls Friday forward three days", func() {
		d := time.Date(2020, 3, 6, 0, 0, 0, 0, time.UTC) // Friday
		Expect(call(fn, d)).To(Equal(time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)))
	})

	It("leaves Saturday, Sunday, and Monday unchanged", func() {
		sat := time.Date(2020, 3, 7, 0, 0, 0, 0, time.UTC)
		sun := time.Date(2020, 3, 8, 0, 0, 0, 0, time.UTC)
		mon := time.Date(2020, 3, 9, 0, 0, 0, 0, time.UTC)
		Expect(call(fn, sat)).To(Equal(sat))
		Expect(call(fn, sun)).To(Equal(sun))
		Expect(call(fn, mon)).To(Equal(mon))
	})
})
