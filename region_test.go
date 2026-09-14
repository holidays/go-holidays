package holidays_test

import (
	"time"

	holidays "github.com/holidays/go-holidays"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Hand-written black-box tests for region-scoped behavior: the RegionName /
// RegionNames lookup API, and regression tests for individual per-country
// observance methods whose weekday or lunar edge cases came out of a specific
// parity divergence or upstream bump. Each method block locks one method's
// behavior so a future refactor cannot silently regress it. The generated table
// tests in internal/definitions cover the broad date corpus; these pin the
// tricky spots.

var _ = Describe("RegionName / RegionNames", func() {
	It("returns the display name for a known region", func() {
		name, ok := holidays.RegionName("gb")
		Expect(ok).To(BeTrue())
		Expect(name).To(Equal("United Kingdom"))
	})

	It("returns comma-ok false for an unknown region", func() {
		name, ok := holidays.RegionName("does_not_exist")
		Expect(ok).To(BeFalse())
		Expect(name).To(Equal(""))
	})

	It("round-trips a non-ASCII display name", func() {
		name, ok := holidays.RegionName("ch_ge")
		Expect(ok).To(BeTrue())
		Expect(name).To(Equal("Genève"))
	})

	It("returns every registered region with the expected count", func() {
		names := holidays.RegionNames()
		Expect(names).To(HaveLen(290))
		Expect(names).To(HaveKeyWithValue("gb", "United Kingdom"))
		Expect(names).To(HaveKeyWithValue("ch_ge", "Genève"))
	})
})

// mustOn resolves holidays for a single date and region with observed dates on.
func mustOn(dateStr, region string) []holidays.Holiday {
	d, err := time.Parse("2006-01-02", dateStr)
	Expect(err).NotTo(HaveOccurred())
	hs, err := holidays.On(d, holidays.Options{Regions: []string{region}, Observed: true})
	Expect(err).NotTo(HaveOccurred())
	return hs
}

// assertHeroesDay checks whether "National Heroes Day" resolves on the given
// date for the ph region, asserting presence (want=true) or absence (want=false).
func assertHeroesDay(opts holidays.Options, date time.Time, want bool) {
	hs, err := holidays.On(date, opts)
	Expect(err).NotTo(HaveOccurred())
	Expect(hasNamedHoliday(hs, "National Heroes Day")).To(Equal(want),
		"National Heroes Day on %s: want present=%v", date.Format("2006-01-02"), want)
}

// Seollal eve (설날 연휴) is the day before Seollal, the first day of the first
// lunar month. Upstream kr.yaml models it as kr_seollal_eve(year, region),
// which is lunar_to_solar(year, 1, 1, region) minus one day.
var _ = Describe("KR Seollal eve", func() {
	DescribeTable("falls on the day before Seollal",
		func(dateStr string) {
			d, err := time.Parse("2006-01-02", dateStr)
			Expect(err).NotTo(HaveOccurred())
			hs, err := holidays.On(d, holidays.Options{Regions: []string{"kr"}, Informal: true})
			Expect(err).NotTo(HaveOccurred())
			Expect(hasNamedHoliday(hs, "설날 연휴")).To(BeTrue(),
				"expected Seollal eve on %s", dateStr)
		},
		Entry("2017", "2017-01-27"),
		Entry("2020", "2020-01-24"),
		Entry("2022", "2022-01-31"),
		Entry("2025", "2025-01-28"),
	)
})

// go-holidays-coi: the two AU year-based Boxing/Proclamation observance methods
// and their expected observed dates across every Dec-26 weekday.
//
//	Boxing Day (au_tas, au_nt) uses to_weekday_if_boxing_weekend_from_year,
//	defined as to_tuesday_if_sunday_or_monday_if_saturday(Dec26):
//	Sat->+2, Sun->+2, Mon UNCHANGED, otherwise unchanged.
//
//	Proclamation Day (au_sa) uses ..._or_to_tuesday_if_monday,
//	defined as to_weekday_if_boxing_weekend(Dec26):
//	Sat->+2, Sun->+2, Mon->+1, otherwise unchanged.
var _ = Describe("AU Boxing/Proclamation Day observance", func() {
	It("matches the gem for Boxing Day (au_tas, au_nt) across every Dec-26 weekday", func() {
		// date string -> expected observed Boxing Day date, keyed by Dec-26 weekday.
		cases := []struct {
			year int
			wday string
			want string // observed Boxing Day date
		}{
			{1970, "Sat", "1970-12-28"},
			{1971, "Sun", "1971-12-28"},
			{1972, "Tue", "1972-12-26"},
			{1973, "Wed", "1973-12-26"},
			{1974, "Thu", "1974-12-26"},
			{1975, "Fri", "1975-12-26"},
			{1977, "Mon", "1977-12-26"}, // the bug: Go currently emits 1977-12-27
		}
		for _, region := range []string{"au_tas", "au_nt"} {
			for _, c := range cases {
				hs := mustOn(c.want, region)
				Expect(hasNamedHoliday(hs, "Boxing Day")).To(BeTrue(),
					"%s Dec26=%s: Boxing Day not observed on %s (want it here)", region, c.wday, c.want)
			}
		}
	})

	It("matches the gem for Proclamation Day (au_sa) across every Dec-26 weekday", func() {
		cases := []struct {
			year int
			wday string
			want string // observed Proclamation Day date
		}{
			{1970, "Sat", "1970-12-28"},
			{1971, "Sun", "1971-12-28"}, // the bug: Go currently emits 1971-12-27
			{1972, "Tue", "1972-12-26"},
			{1973, "Wed", "1973-12-26"},
			{1974, "Thu", "1974-12-26"},
			{1975, "Fri", "1975-12-26"},
			{1977, "Mon", "1977-12-27"},
		}
		for _, c := range cases {
			hs := mustOn(c.want, "au_sa")
			Expect(hasNamedHoliday(hs, "Proclamation Day")).To(BeTrue(),
				"au_sa Dec26=%s: Proclamation Day not observed on %s (want it here)", c.wday, c.want)
		}
	})
})

// ph National Heroes Day is the last Monday of August. This is the engine's
// authoritative behavior and is asserted here directly. Upstream ph.yaml used to
// carry an off-by-one (holidays/definitions#345) that emitted September 1 in
// years where August 31 is a Sunday; this engine always refused to reproduce
// that quirk, which made it the parity sweep's last known divergence. The fix
// shipped in definitions v8.0.2 as function: ph_heroes_day(year), so the sweep
// allowlist is gone. These cases stay to lock the behavior so a future refactor
// cannot silently regress it to the buggy date.
var _ = Describe("PH National Heroes Day", func() {
	It("falls on the last Monday of August", func() {
		opts := holidays.Options{Regions: []string{"ph"}}

		// The eight years in 1970-2050 where August 31 falls on a Sunday: the
		// correct last Monday of August is the 25th (the buggy off-by-one rule
		// yields September 1 for exactly these years).
		sundayAug31Years := []int{1975, 1980, 1986, 1997, 2003, 2008, 2014, 2025}
		for _, y := range sundayAug31Years {
			assertHeroesDay(opts, time.Date(y, time.August, 25, 0, 0, 0, 0, time.UTC), true)
			assertHeroesDay(opts, time.Date(y, time.September, 1, 0, 0, 0, 0, time.UTC), false)
		}

		// Control years where August 31 is not a Sunday: last Monday of August is
		// unambiguous.
		controls := map[int]int{2020: 31, 2021: 30, 2022: 29, 2023: 28, 2024: 26}
		for y, mday := range controls {
			assertHeroesDay(opts, time.Date(y, time.August, mday, 0, 0, 0, 0, time.UTC), true)
		}
	})
})
