package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ie_st_brigids_day", func() {
	call := func(year int) time.Time {
		fn, ok := LookupMethod("ie_st_brigids_day")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("returns February 1 itself when it falls on a Friday", func() {
		// February 1, 2019 is a Friday.
		Expect(call(2019)).To(Equal(time.Date(2019, 2, 1, 0, 0, 0, 0, time.UTC)))
	})

	It("otherwise returns the first Monday in February", func() {
		// February 1, 2020 is a Saturday.
		Expect(call(2020)).To(Equal(time.Date(2020, 2, 3, 0, 0, 0, 0, time.UTC)))
	})

	It("returns February 1 itself when it already falls on a Monday", func() {
		// February 1, 2021 is a Monday.
		Expect(call(2021)).To(Equal(time.Date(2021, 2, 1, 0, 0, 0, 0, time.UTC)))
	})
})
