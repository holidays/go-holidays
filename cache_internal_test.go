package holidays

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Unit test for the cache internals (see the test taxonomy in
// .claude/rules/project-conventions.md): sentinel data only, no real definition
// content. It is white-box (`package holidays`, not `holidays_test`) so it can
// reach the unexported store, and it lives at the module root only because
// cache.go does. This file's Describe/It blocks register into the same
// process-global Ginkgo spec tree as the holidays_test package; the single
// RunSpecs bootstrap for the whole root test binary lives in
// holidays_suite_test.go. Do not add another RunSpecs call here.

var _ = Describe("cache internals", func() {
	BeforeEach(func() {
		ResetCache()
	})
	AfterEach(func() {
		ResetCache()
	})

	It("stores and finds an exact range", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		opts := Options{Regions: []string{"us"}}
		sentinel := []Holiday{
			{Date: time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC), Name: "Test 4", Regions: []string{"us"}},
			{Date: time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Test 25", Regions: []string{"us"}},
		}
		cacheStore(from, to, opts, sentinel)

		got, ok := cacheFind(from, to, opts)
		Expect(ok).To(BeTrue(), "expected cache hit on exact range")
		Expect(got).To(HaveLen(2))
	})

	It("hits the cache for a range contained within the stored range", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		opts := Options{Regions: []string{"us"}}
		sentinel := []Holiday{
			{Date: time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC), Name: "Test 4"},
			{Date: time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC), Name: "Test 25"},
		}
		cacheStore(from, to, opts, sentinel)

		// Narrower range entirely inside the cached one: should still hit.
		got, ok := cacheFind(
			time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC),
			opts,
		)
		Expect(ok).To(BeTrue(), "expected cache hit on contained range")
		Expect(got).To(HaveLen(1))
		Expect(got[0].Name).To(Equal("Test 4"))
	})

	It("misses when the requested range extends outside the cached range", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		opts := Options{Regions: []string{"us"}}
		cacheStore(from, to, opts, nil)

		// Request range extending past cached upper bound.
		_, ok := cacheFind(
			time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2027, 1, 31, 0, 0, 0, 0, time.UTC),
			opts,
		)
		Expect(ok).To(BeFalse(), "expected cache miss for range extending past cache")

		// Request range starting before cached lower bound.
		_, ok = cacheFind(
			time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC),
			opts,
		)
		Expect(ok).To(BeFalse(), "expected cache miss for range starting before cache")
	})

	It("misses when options differ", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		cacheStore(from, to, Options{Regions: []string{"us"}}, nil)

		_, ok := cacheFind(from, to, Options{Regions: []string{"gb"}})
		Expect(ok).To(BeFalse(), "different region should miss")

		_, ok = cacheFind(from, to, Options{Regions: []string{"us"}, Informal: true})
		Expect(ok).To(BeFalse(), "flipping informal should miss")

		_, ok = cacheFind(from, to, Options{Regions: []string{"us"}, Observed: true})
		Expect(ok).To(BeFalse(), "flipping observed should miss")
	})

	It("breaks a same-date tie by name ascending", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		opts := Options{Regions: []string{"us"}}
		sameDay := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
		sentinel := []Holiday{
			{Date: sameDay, Name: "Zeta Day"},
			{Date: sameDay, Name: "Alpha Day"},
		}
		cacheStore(from, to, opts, sentinel)

		got, ok := cacheFind(from, to, opts)
		Expect(ok).To(BeTrue())
		Expect(got).To(HaveLen(2))
		Expect(got[0].Name).To(Equal("Alpha Day"), "expected same-date entries sorted by name ascending")
		Expect(got[1].Name).To(Equal("Zeta Day"))
	})

	It("computes an options key that is order-insensitive for regions", func() {
		a := optionsKey(Options{Regions: []string{"us", "gb"}})
		b := optionsKey(Options{Regions: []string{"gb", "us"}})
		Expect(a).To(Equal(b))
	})

	It("clears the cache on reset", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		opts := Options{Regions: []string{"us"}}
		cacheStore(from, to, opts, []Holiday{{Date: from, Name: "x"}})

		_, ok := cacheFind(from, to, opts)
		Expect(ok).To(BeTrue(), "expected hit before reset")

		ResetCache()

		_, ok = cacheFind(from, to, opts)
		Expect(ok).To(BeFalse(), "expected miss after reset")
	})
})
