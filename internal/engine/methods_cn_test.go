package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("cn_qingming", func() {
	It("returns the Qingming solar term date for a supported year", func() {
		fn, ok := LookupMethod("cn_qingming")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: 2020})
		Expect(err).NotTo(HaveOccurred())
		Expect(got.Month()).To(Equal(time.April))
	})

	It("propagates an error for a year outside the supported range", func() {
		fn, ok := LookupMethod("cn_qingming")
		Expect(ok).To(BeTrue())
		_, err := fn(MethodArgs{Year: 1800})
		Expect(err).To(HaveOccurred())
	})
})
