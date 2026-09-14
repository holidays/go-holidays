package engine

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/holidays/go-holidays/internal/definition"
)

// Regression coverage for go-holidays-dpt: a multi-segment wildcard region
// (e.g. "au_vic_") must collapse all the way down to the country segment,
// matching the same rules as the single-segment wildcard ("au_"), not just
// the rules whose region starts with the full multi-segment prefix.
var _ = Describe("ruleMatchesRequested", func() {
	var (
		countryWide definition.HolidayRule
		midLevel    definition.HolidayRule
		leaf        definition.HolidayRule
	)

	BeforeEach(func() {
		countryWide = definition.HolidayRule{Name: "Country Wide Day", Regions: []string{"au"}}
		midLevel = definition.HolidayRule{Name: "Victoria Day", Regions: []string{"au_vic"}}
		leaf = definition.HolidayRule{Name: "Melbourne Cup", Regions: []string{"au_vic_melbourne"}}
	})

	Context("with a multi-segment wildcard request (au_vic_)", func() {
		requested := []string{"au_vic_"}

		It("matches a country-wide-only rule", func() {
			Expect(ruleMatchesRequested(countryWide, requested)).To(BeTrue())
		})

		It("matches a mid-level rule under the same branch", func() {
			Expect(ruleMatchesRequested(midLevel, requested)).To(BeTrue())
		})

		It("matches a leaf rule under the same branch", func() {
			Expect(ruleMatchesRequested(leaf, requested)).To(BeTrue())
		})

		It("matches a rule under a sibling branch, identically to the single-segment wildcard", func() {
			sibling := definition.HolidayRule{Name: "NSW Day", Regions: []string{"au_nsw"}}
			Expect(ruleMatchesRequested(sibling, requested)).To(BeTrue())
			Expect(ruleMatchesRequested(sibling, []string{"au_"})).To(BeTrue())
		})
	})

	Context("with a single-segment wildcard request (au_)", func() {
		requested := []string{"au_"}

		It("still matches every level (no regression)", func() {
			Expect(ruleMatchesRequested(countryWide, requested)).To(BeTrue())
			Expect(ruleMatchesRequested(midLevel, requested)).To(BeTrue())
			Expect(ruleMatchesRequested(leaf, requested)).To(BeTrue())
		})
	})

	Context("with an unrelated wildcard request", func() {
		It("does not match a different country's rule", func() {
			usRule := definition.HolidayRule{Name: "US Day", Regions: []string{"us_ga"}}
			Expect(ruleMatchesRequested(usRule, []string{"au_vic_"})).To(BeFalse())
		})
	})
})
