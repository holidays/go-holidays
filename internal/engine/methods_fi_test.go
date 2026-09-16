package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("fi methods", func() {
	call := func(name string, year int) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Describe("fi_juhannusaatto", func() {
		It("returns June 25 when June 19 is a Saturday", func() {
			Expect(call("fi_juhannusaatto", 2021)).To(Equal(time.Date(2021, 6, 25, 0, 0, 0, 0, time.UTC)))
		})

		It("returns the Friday between June 19 and 25 otherwise", func() {
			// June 19, 2023 is a Monday.
			Expect(call("fi_juhannusaatto", 2023)).To(Equal(time.Date(2023, 6, 23, 0, 0, 0, 0, time.UTC)))
		})
	})

	It("fi_juhannuspaiva is the first Saturday on or after June 20", func() {
		got := call("fi_juhannuspaiva", 2020)
		Expect(got.Weekday()).To(Equal(time.Saturday))
		Expect(got.Before(time.Date(2020, 6, 27, 0, 0, 0, 0, time.UTC))).To(BeTrue())
	})

	It("fi_pyhainpaiva is the first Saturday on or after October 31", func() {
		got := call("fi_pyhainpaiva", 2020)
		Expect(got.Weekday()).To(Equal(time.Saturday))
		Expect(got.Month()).To(BeElementOf(time.October, time.November))
	})
})
