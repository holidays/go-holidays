package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("solarTermDay", func() {
	It("errors for a year outside every supported base range", func() {
		_, err := solarTermDay(1700, time.March, vernalEquinoxBase)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("out of supported range"))
	})

	It("computes a date within each defined base range", func() {
		for _, year := range []int{1860, 1920, 2000, 2120} {
			got, err := solarTermDay(year, time.March, vernalEquinoxBase)
			Expect(err).NotTo(HaveOccurred())
			Expect(got.Year()).To(Equal(year))
		}
	})
})
