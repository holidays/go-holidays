package engine

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/holidays/go-holidays/internal/definition"
)

// Regression coverage for go-holidays-67s: two definitions that share every
// field compared for identity (name, wday, mday, week, function,
// function_modifier, type, observed, year_ranges) but fall in different
// months are distinct holidays. Definitions are bucketed per month, so month
// is implicitly part of a definition's identity. ruleSignature must therefore
// include the month, or ResolveYear's dedup silently drops the second one on an
// aggregate query (e.g. "au_" losing the ACT/NSW/SA October Labour Day because
// it collides with the WA March Labour Day).
var _ = Describe("ruleSignature month component", func() {
	AfterEach(func() {
		UnregisterCountry("zz")
	})

	It("keeps two same-named first-Monday rules that differ only by month", func() {
		RegisterCountry("zz", []definition.HolidayRule{
			{Name: "Labour Day", Regions: []string{"zz_west"}, Month: 3, Wday: 1, Week: 1},
			{Name: "Labour Day", Regions: []string{"zz_east"}, Month: 10, Wday: 1, Week: 1},
		})

		got, err := ResolveYear(2020, ResolveOptions{Regions: []string{"zz_"}})
		Expect(err).NotTo(HaveOccurred())

		var months []int
		for _, r := range got {
			if r.Name == "Labour Day" {
				months = append(months, int(r.Date.Month()))
			}
		}
		Expect(months).To(ConsistOf(3, 10),
			"expected both the March and October Labour Day, got %+v", got)
	})
})
