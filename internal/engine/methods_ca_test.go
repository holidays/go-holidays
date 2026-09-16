package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ca_victoria_day", func() {
	It("rolls back to the Monday before May 24 when May 24 is not a Monday", func() {
		// May 24, 2022 is a Tuesday.
		fn, ok := LookupMethod("ca_victoria_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2022})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2022, 5, 23, 0, 0, 0, 0, time.UTC)))
	})

	It("returns May 24 itself when it already falls on a Monday", func() {
		// May 24, 2021 is a Monday.
		fn, ok := LookupMethod("ca_victoria_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2021})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(time.Date(2021, 5, 24, 0, 0, 0, 0, time.UTC)))
	})
})
