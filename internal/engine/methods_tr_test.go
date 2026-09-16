package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("tr methods", func() {
	call := func(name string, year int) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("ramadan_feast returns the proclaimed date for a known year", func() {
		Expect(call("ramadan_feast", 2020)).To(Equal(time.Date(2020, time.May, 24, 0, 0, 0, 0, time.UTC)))
	})

	It("ramadan_feast returns zero for a year outside the table", func() {
		Expect(call("ramadan_feast", 1999).IsZero()).To(BeTrue())
	})

	It("sacrifice_feast returns the proclaimed date for a known year", func() {
		Expect(call("sacrifice_feast", 2020)).To(Equal(time.Date(2020, time.July, 31, 0, 0, 0, 0, time.UTC)))
	})

	It("sacrifice_feast returns zero for a year outside the table", func() {
		Expect(call("sacrifice_feast", 1999).IsZero()).To(BeTrue())
	})
})
