package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Regression coverage for independence_day: definitions master (commit
// 4488c2f) added this as the observed-shift function for US Independence
// Day, replacing the old us_va-only to_weekday_if_weekend hack. Behavior
// varies by region: us_ri shifts forward to the next Monday on a weekend,
// us_tx never shifts, and everyone else uses nearest-weekday.
var _ = Describe("independence_day", func() {
	var fn Method

	BeforeEach(func() {
		var ok bool
		fn, ok = LookupMethod("independence_day")
		Expect(ok).To(BeTrue(), "independence_day must be registered")
	})

	call := func(region string, date time.Time) time.Time {
		got, err := fn(MethodArgs{Region: region, Date: date})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	Context("default regions (us, us_va, ...)", func() {
		It("shifts a Saturday back to Friday", func() {
			Expect(call("us", time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC))).
				To(Equal(time.Date(2026, 7, 3, 0, 0, 0, 0, time.UTC)))
		})

		It("shifts a Sunday forward to Monday", func() {
			Expect(call("us_va", time.Date(2027, 7, 4, 0, 0, 0, 0, time.UTC))).
				To(Equal(time.Date(2027, 7, 5, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a weekday unchanged", func() {
			d := time.Date(2028, 7, 4, 0, 0, 0, 0, time.UTC) // Tuesday
			Expect(call("us", d)).To(Equal(d))
		})
	})

	Context("us_ri", func() {
		It("shifts a Saturday forward two days to Monday", func() {
			Expect(call("us_ri", time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC))).
				To(Equal(time.Date(2026, 7, 6, 0, 0, 0, 0, time.UTC)))
		})

		It("shifts a Sunday forward one day to Monday", func() {
			Expect(call("us_ri", time.Date(2027, 7, 4, 0, 0, 0, 0, time.UTC))).
				To(Equal(time.Date(2027, 7, 5, 0, 0, 0, 0, time.UTC)))
		})

		It("leaves a weekday unchanged", func() {
			d := time.Date(2028, 7, 4, 0, 0, 0, 0, time.UTC)
			Expect(call("us_ri", d)).To(Equal(d))
		})
	})

	Context("us_tx", func() {
		It("never shifts, even on a weekend", func() {
			d := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC) // Saturday
			Expect(call("us_tx", d)).To(Equal(d))
		})
	})
})
