package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("is_sumardagurinn_fyrsti", func() {
	call := func(year int) time.Time {
		fn, ok := LookupMethod("is_sumardagurinn_fyrsti")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("returns April 25 when April 18 is itself a Thursday", func() {
		// April 18, 2019 is a Thursday.
		Expect(call(2019)).To(Equal(time.Date(2019, 4, 25, 0, 0, 0, 0, time.UTC)))
	})

	It("returns the Thursday strictly after April 18 when it falls earlier in the week", func() {
		// April 18, 2018 is a Wednesday.
		Expect(call(2018)).To(Equal(time.Date(2018, 4, 19, 0, 0, 0, 0, time.UTC)))
	})
})
