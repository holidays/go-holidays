package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ch methods", func() {
	call := func(name string, args MethodArgs) time.Time {
		fn, ok := LookupMethod(name)
		Expect(ok).To(BeTrue(), "method %q must be registered", name)
		got, err := fn(args)
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("ch_vd_lundi_du_jeune_federal is the Monday after the third Sunday of September", func() {
		got := call("ch_vd_lundi_du_jeune_federal", MethodArgs{Year: 2020})
		Expect(got.Weekday()).To(Equal(time.Monday))
	})

	It("ch_ge_jeune_genevois is the Thursday after the first Sunday of September", func() {
		got := call("ch_ge_jeune_genevois", MethodArgs{Year: 2020})
		Expect(got.Weekday()).To(Equal(time.Thursday))
	})

	It("ch_be_zibelemaerit is the fourth Monday of November", func() {
		got := call("ch_be_zibelemaerit", MethodArgs{Year: 2020})
		Expect(got.Weekday()).To(Equal(time.Monday))
		Expect(got.Month()).To(Equal(time.November))
	})

	Describe("ch_gl_naefelser_fahrt", func() {
		It("shifts by a week when the first Thursday of April is Maundy Thursday", func() {
			// 2015: first Thursday of April (Apr 2) coincides with Maundy Thursday.
			got := call("ch_gl_naefelser_fahrt", MethodArgs{Year: 2015})
			Expect(got).To(Equal(time.Date(2015, 4, 9, 0, 0, 0, 0, time.UTC)))
		})

		It("stays on the first Thursday of April otherwise", func() {
			got := call("ch_gl_naefelser_fahrt", MethodArgs{Year: 2016})
			Expect(got).To(Equal(time.Date(2016, 4, 7, 0, 0, 0, 0, time.UTC)))
		})
	})
})
