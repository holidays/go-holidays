package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("de_buss_und_bettag", func() {
	call := func(year int) time.Time {
		fn, ok := LookupMethod("de_buss_und_bettag")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("returns November 16 when November 23 is a Wednesday", func() {
		// November 23, 2016 is a Wednesday.
		Expect(call(2016)).To(Equal(time.Date(2016, 11, 16, 0, 0, 0, 0, time.UTC)))
	})

	It("returns the Wednesday before, when November 23 is after Wednesday in the week", func() {
		// November 23, 2018 is a Friday (weekday index 5, > 3).
		Expect(call(2018)).To(Equal(time.Date(2018, 11, 21, 0, 0, 0, 0, time.UTC)))
	})

	It("returns the Wednesday before, when November 23 is on/before Wednesday in the week", func() {
		// November 23, 2020 is a Monday (weekday index 1, <= 3).
		Expect(call(2020)).To(Equal(time.Date(2020, 11, 18, 0, 0, 0, 0, time.UTC)))
	})
})
