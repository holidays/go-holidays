package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("se methods", func() {
	call := func(name string, year int) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("se_midsommardagen is the first Saturday on or after June 20", func() {
		got := call("se_midsommardagen", 2020)
		Expect(got.Weekday()).To(Equal(time.Saturday))
		Expect(got.Month()).To(Equal(time.June))
		Expect(got.Day()).To(BeNumerically(">=", 20))
	})

	It("se_alla_helgons_dag is the first Saturday on or after October 31", func() {
		got := call("se_alla_helgons_dag", 2020)
		Expect(got.Weekday()).To(Equal(time.Saturday))
	})

	It("nextSaturdayOnOrAfter returns the anchor day itself when it is already a Saturday", func() {
		// June 20, 2020 is a Saturday.
		Expect(nextSaturdayOnOrAfter(2020, time.June, 20)).To(Equal(time.Date(2020, 6, 20, 0, 0, 0, 0, time.UTC)))
	})
})
