package engine

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("kr_seollal_eve", func() {
	It("returns the day before Seollal (lunar new year) for a supported region/year", func() {
		fn, ok := LookupMethod("kr_seollal_eve")
		Expect(ok).To(BeTrue())
		seollalFn, ok := LookupMethod("lunar_to_solar")
		Expect(ok).To(BeTrue())

		seollal, err := seollalFn(MethodArgs{Year: 2021, Month: 1, Day: 1, Region: "kr"})
		Expect(err).NotTo(HaveOccurred())

		got, err := fn(MethodArgs{Year: 2021, Region: "kr"})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(seollal.AddDate(0, 0, -1)))
	})

	It("propagates an error for an unsupported region", func() {
		fn, ok := LookupMethod("kr_seollal_eve")
		Expect(ok).To(BeTrue())
		_, err := fn(MethodArgs{Year: 2021, Region: "zz_not_lunar"})
		Expect(err).To(HaveOccurred())
	})
})
