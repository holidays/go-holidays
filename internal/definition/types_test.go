package definition_test

import (
	"github.com/holidays/go-holidays/internal/definition"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HolidayRule predicates", func() {
	DescribeTable("HasWday",
		func(week int, want bool) {
			r := definition.HolidayRule{Week: week}
			Expect(r.HasWday()).To(Equal(want))
		},
		Entry("nonzero week", 2, true),
		Entry("zero week", 0, false),
	)

	DescribeTable("HasMday",
		func(mday int, want bool) {
			r := definition.HolidayRule{Mday: mday}
			Expect(r.HasMday()).To(Equal(want))
		},
		Entry("nonzero mday", 25, true),
		Entry("zero mday", 0, false),
	)

	DescribeTable("HasFunction",
		func(fn string, want bool) {
			r := definition.HolidayRule{Function: fn}
			Expect(r.HasFunction()).To(Equal(want))
		},
		Entry("nonempty function", "easter", true),
		Entry("empty function", "", false),
	)

	DescribeTable("HasObserved",
		func(observed string, want bool) {
			r := definition.HolidayRule{Observed: observed}
			Expect(r.HasObserved()).To(Equal(want))
		},
		Entry("nonempty observed", "to_monday_if_sunday", true),
		Entry("empty observed", "", false),
	)
})

var _ = Describe("HolidayRule.AppliesIn", func() {
	It("applies in every year when YearRanges is empty", func() {
		r := definition.HolidayRule{}
		Expect(r.AppliesIn(1999)).To(BeTrue())
		Expect(r.AppliesIn(2050)).To(BeTrue())
	})

	It("applies when at least one YearRange matches", func() {
		r := definition.HolidayRule{
			YearRanges: []definition.YearRange{
				{Kind: definition.YearRangeLimited, Years: []int{1999}},
				{Kind: definition.YearRangeFrom, Years: []int{2020}},
			},
		}
		Expect(r.AppliesIn(2020)).To(BeTrue())
	})

	It("does not apply when no YearRange matches", func() {
		r := definition.HolidayRule{
			YearRanges: []definition.YearRange{
				{Kind: definition.YearRangeLimited, Years: []int{1999}},
				{Kind: definition.YearRangeFrom, Years: []int{2020}},
			},
		}
		Expect(r.AppliesIn(2019)).To(BeFalse())
	})
})

var _ = Describe("YearRange.Matches", func() {
	Context("YearRangeAny", func() {
		It("matches every year", func() {
			yr := definition.YearRange{Kind: definition.YearRangeAny}
			Expect(yr.Matches(1900)).To(BeTrue())
			Expect(yr.Matches(2100)).To(BeTrue())
		})
	})

	Context("YearRangeUntil", func() {
		yr := definition.YearRange{Kind: definition.YearRangeUntil, Years: []int{2000}}

		DescribeTable("matches years up to and including the bound",
			func(year int, want bool) {
				Expect(yr.Matches(year)).To(Equal(want))
			},
			Entry("year before bound", 1999, true),
			Entry("year equal to bound", 2000, true),
			Entry("year after bound", 2001, false),
		)

		It("does not match when Years does not contain exactly one value", func() {
			Expect(definition.YearRange{Kind: definition.YearRangeUntil, Years: []int{}}.Matches(1999)).To(BeFalse())
			Expect(definition.YearRange{Kind: definition.YearRangeUntil, Years: []int{2000, 2001}}.Matches(1999)).To(BeFalse())
		})
	})

	Context("YearRangeFrom", func() {
		yr := definition.YearRange{Kind: definition.YearRangeFrom, Years: []int{2000}}

		DescribeTable("matches years from the bound onward",
			func(year int, want bool) {
				Expect(yr.Matches(year)).To(Equal(want))
			},
			Entry("year before bound", 1999, false),
			Entry("year equal to bound", 2000, true),
			Entry("year after bound", 2001, true),
		)

		It("does not match when Years does not contain exactly one value", func() {
			Expect(definition.YearRange{Kind: definition.YearRangeFrom, Years: []int{}}.Matches(2000)).To(BeFalse())
			Expect(definition.YearRange{Kind: definition.YearRangeFrom, Years: []int{2000, 2001}}.Matches(2000)).To(BeFalse())
		})
	})

	Context("YearRangeLimited", func() {
		yr := definition.YearRange{Kind: definition.YearRangeLimited, Years: []int{1999, 2003, 2010}}

		DescribeTable("matches only years present in the set",
			func(year int, want bool) {
				Expect(yr.Matches(year)).To(Equal(want))
			},
			Entry("first listed year", 1999, true),
			Entry("middle listed year", 2003, true),
			Entry("last listed year", 2010, true),
			Entry("year not in the set", 2004, false),
		)

		It("matches nothing when Years is empty", func() {
			Expect(definition.YearRange{Kind: definition.YearRangeLimited}.Matches(2000)).To(BeFalse())
		})
	})

	Context("YearRangeBetween", func() {
		yr := definition.YearRange{Kind: definition.YearRangeBetween, Years: []int{2000, 2010}}

		DescribeTable("matches years within the inclusive span",
			func(year int, want bool) {
				Expect(yr.Matches(year)).To(Equal(want))
			},
			Entry("year before span", 1999, false),
			Entry("start of span", 2000, true),
			Entry("inside span", 2005, true),
			Entry("end of span", 2010, true),
			Entry("year after span", 2011, false),
		)

		It("does not match when Years does not contain exactly two values", func() {
			Expect(definition.YearRange{Kind: definition.YearRangeBetween, Years: []int{2000}}.Matches(2000)).To(BeFalse())
			Expect(definition.YearRange{Kind: definition.YearRangeBetween, Years: []int{}}.Matches(2000)).To(BeFalse())
			Expect(definition.YearRange{Kind: definition.YearRangeBetween, Years: []int{2000, 2005, 2010}}.Matches(2000)).To(BeFalse())
		})
	})

	Context("an unrecognized Kind", func() {
		It("falls back to matching every year", func() {
			yr := definition.YearRange{Kind: definition.YearRangeKind(99)}
			Expect(yr.Matches(1900)).To(BeTrue())
			Expect(yr.Matches(2100)).To(BeTrue())
		})
	})
})
