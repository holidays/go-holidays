package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/holidays/go-holidays/internal/definition"
)

var _ = Describe("RegisterMethod", func() {
	It("registers a new method under a unique name", func() {
		RegisterMethod("zz_test_register_method_once", func(a MethodArgs) (time.Time, error) {
			return a.Date, nil
		})
		_, ok := LookupMethod("zz_test_register_method_once")
		Expect(ok).To(BeTrue())
	})

	It("panics when the name is already registered", func() {
		RegisterMethod("zz_test_register_method_dup", func(a MethodArgs) (time.Time, error) {
			return a.Date, nil
		})
		Expect(func() {
			RegisterMethod("zz_test_register_method_dup", func(a MethodArgs) (time.Time, error) {
				return a.Date, nil
			})
		}).To(PanicWith(ContainSubstring("already registered")))
	})
})

var _ = Describe("IsMethodRegistered", func() {
	It("returns true for a registered method", func() {
		RegisterMethod("zz_test_is_registered", func(a MethodArgs) (time.Time, error) {
			return a.Date, nil
		})
		Expect(IsMethodRegistered("zz_test_is_registered")).To(BeTrue())
	})

	It("returns false for an unregistered method", func() {
		Expect(IsMethodRegistered("zz_test_not_registered_at_all")).To(BeFalse())
	})
})

var _ = Describe("AvailableRegions", func() {
	AfterEach(func() {
		UnregisterCountry("zzavail")
	})

	It("returns every region code mentioned across registered rules, sorted", func() {
		RegisterCountry("zzavail", []definition.HolidayRule{
			{Name: "A", Regions: []string{"zzavail_b", "zzavail_a"}},
			{Name: "B", Regions: []string{"zzavail_c"}},
		})
		regions := AvailableRegions()
		Expect(regions).To(ContainElements("zzavail_a", "zzavail_b", "zzavail_c"))
		// Sorted check: find the indices among the filtered subset.
		var got []string
		for _, r := range regions {
			if r == "zzavail_a" || r == "zzavail_b" || r == "zzavail_c" {
				got = append(got, r)
			}
		}
		Expect(got).To(Equal([]string{"zzavail_a", "zzavail_b", "zzavail_c"}))
	})
})

var _ = Describe("rulesForCountry", func() {
	AfterEach(func() {
		UnregisterCountry("zzrfc")
	})

	It("returns the rules registered for that country", func() {
		rules := []definition.HolidayRule{{Name: "Only Rule", Regions: []string{"zzrfc"}}}
		RegisterCountry("zzrfc", rules)
		Expect(rulesForCountry("zzrfc")).To(Equal(rules))
	})

	It("returns nil for an unregistered country", func() {
		Expect(rulesForCountry("zzrfc_does_not_exist")).To(BeNil())
	})
})

var _ = Describe("rulesFor cache", func() {
	AfterEach(func() {
		UnregisterCountry("zzcache")
	})

	It("rebuilds after invalidation and serves the cached slice on subsequent calls", func() {
		RegisterCountry("zzcache", []definition.HolidayRule{{Name: "Cached", Regions: []string{"zzcache"}}})
		first := rulesFor(nil)
		second := rulesFor(nil)
		Expect(first).To(Equal(second))
	})

	It("handles concurrent rebuild races (double-checked locking path)", func() {
		// Deterministically reproduce the interleaving the double check
		// guards against: caller A observes the cache as invalid under a
		// read lock and releases it; before A acquires the write lock,
		// caller B (running concurrently) rebuilds the cache and marks it
		// valid; A must then see the now-valid cache once it finally
		// acquires the write lock, rather than rebuilding again.
		RegisterCountry("zzcache", []definition.HolidayRule{{Name: "Cached", Regions: []string{"zzcache"}}})

		release := make(chan struct{})
		afterInvalidRead = func() {
			afterInvalidRead = nil // only the first (A's) read pauses
			go func() {
				rulesFor(nil) // B: rebuilds and marks the cache valid
				close(release)
			}()
			<-release // wait for B to finish before A takes the write lock
		}
		defer func() { afterInvalidRead = nil }()

		rulesFor(nil) // A: hits the double check once it acquires the write lock
	})
})

var _ = Describe("ruleMatchesRequested additional branches", func() {
	It("matches everything when no regions are requested", func() {
		rule := definition.HolidayRule{Name: "Any", Regions: []string{"anything"}}
		Expect(ruleMatchesRequested(rule, nil)).To(BeTrue())
	})

	It("matches an exact non-wildcard region", func() {
		rule := definition.HolidayRule{Name: "Exact", Regions: []string{"us_ga"}}
		Expect(ruleMatchesRequested(rule, []string{"us_ga"})).To(BeTrue())
	})

	It("matches via parent for a non-wildcard request", func() {
		rule := definition.HolidayRule{Name: "Parent", Regions: []string{"us"}}
		Expect(ruleMatchesRequested(rule, []string{"us_ga"})).To(BeTrue())
	})

	It("falls through multiple non-matching requested regions before matching one", func() {
		rule := definition.HolidayRule{Name: "Multi", Regions: []string{"us_ga"}}
		Expect(ruleMatchesRequested(rule, []string{"ca_on", "us_ga"})).To(BeTrue())
	})

	It("returns false via the non-wildcard path when nothing matches", func() {
		rule := definition.HolidayRule{Name: "NoMatch", Regions: []string{"us_ga"}}
		Expect(ruleMatchesRequested(rule, []string{"ca_on"})).To(BeFalse())
	})
})

var _ = Describe("isParentOf", func() {
	It("reports true when child is prefixed by parent plus underscore", func() {
		Expect(isParentOf("us", "us_ga")).To(BeTrue())
	})

	It("reports false when child equals parent", func() {
		Expect(isParentOf("us", "us")).To(BeFalse())
	})

	It("reports false when child does not share the parent prefix", func() {
		Expect(isParentOf("us", "ca_on")).To(BeFalse())
	})

	It("reports false when child is shorter than parent", func() {
		Expect(isParentOf("us_ga", "us")).To(BeFalse())
	})
})

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
