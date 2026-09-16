package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ph_heroes_day", func() {
	It("returns the last Monday of August", func() {
		fn, ok := LookupMethod("ph_heroes_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2020})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Weekday()).To(Equal(time.Monday))
		Expect(got.Month()).To(Equal(time.August))
		Expect(got.AddDate(0, 0, 7).Month()).To(Equal(time.September))
	})
})
