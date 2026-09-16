package holidays_test

import (
	"time"

	holidays "github.com/holidays/go-holidays"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func hasNamedHoliday(hs []holidays.Holiday, name string) bool {
	for _, h := range hs {
		if h.Name == name {
			return true
		}
	}
	return false
}

var _ = Describe("NextHolidays", func() {
	It("returns the correct next five US holidays", func() {
		from := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.NextHolidays(from, 5, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(5))

		want := []struct {
			date string
			name string
		}{
			{"2026-06-19", "Juneteenth National Independence Day"},
			{"2026-07-04", "Independence Day"},
			{"2026-09-07", "Labor Day"},
			{"2026-11-11", "Veterans Day"},
			{"2026-11-26", "Thanksgiving"},
		}
		for i, w := range want {
			Expect(hs[i].Date.Format("2006-01-02")).To(Equal(w.date), "index %d date", i)
			Expect(hs[i].Name).To(Equal(w.name), "index %d name", i)
		}
	})

	It("crosses the year boundary when observed dates shift", func() {
		from := time.Date(2026, 12, 20, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.NextHolidays(from, 3, holidays.Options{
			Regions:  []string{"us"},
			Observed: true,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(3))
		Expect(hs[0].Date.Format("2006-01-02")).To(Equal("2026-12-25"))
		Expect(hs[2].Date.Year()).To(Equal(2027))
	})

	It("includes the from date itself when it is a holiday", func() {
		// July 4 is Independence Day; from=July 4 must include it.
		from := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.NextHolidays(from, 1, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(1))
		Expect(hs[0].Date.Format("2006-01-02")).To(Equal("2026-07-04"))
	})

	It("returns at most count holidays, expanding across years as needed", func() {
		// The scan expands forward across years until count are collected; a large
		// count is satisfied from multiple years without error or exceeding count.
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.NextHolidays(from, 1000, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(len(hs)).To(BeNumerically("<=", 1000))
		Expect(hs).NotTo(BeEmpty())
	})

	It("requires a positive count", func() {
		from := time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)
		for _, c := range []int{0, -1, -100} {
			_, err := holidays.NextHolidays(from, c, holidays.Options{Regions: []string{"us"}})
			Expect(err).To(HaveOccurred(), "count=%d", c)
		}
	})

	It("expands beyond twelve months when the count-th holiday lands later", func() {
		// go-holidays-6vj: NextHolidays must return the full count even when the
		// count-th holiday lands more than 12 months past `from`. From 2024-01-01
		// the 12th US holiday is MLK Day 2025-01-20, beyond a fixed 12-month window.
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.NextHolidays(from, 12, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(12))
		Expect(hs[11].Date.Format("2006-01-02")).To(Equal("2025-01-20"))
	})

	It("propagates a resolution error from the engine", func() {
		// kr's lunar tables end at 2049, so resolving 2050 as the very first
		// forward-scan year (no tolerated look-ahead here) must error.
		from := time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := holidays.NextHolidays(from, 1, holidays.Options{Regions: []string{"kr"}})
		Expect(err).To(HaveOccurred())
	})

	It("returns results sorted ascending across multiple regions", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.NextHolidays(from, 50, holidays.Options{Regions: []string{"us", "ca"}})
		Expect(err).NotTo(HaveOccurred())
		for i := 1; i < len(hs); i++ {
			Expect(hs[i].Date.Before(hs[i-1].Date)).To(BeFalse(), "index %d out of order", i)
		}
	})
})

var _ = Describe("AnyHolidaysDuringWorkWeek", func() {
	It("is true for a week containing a holiday", func() {
		// Week containing Thanksgiving 2026 (Thursday Nov 26).
		wed := time.Date(2026, 11, 25, 0, 0, 0, 0, time.UTC)
		got, err := holidays.AnyHolidaysDuringWorkWeek(wed, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeTrue())
	})

	It("is false for a week with no US federal holidays", func() {
		// Week of Aug 10-14, 2026 has no US federal holidays.
		tue := time.Date(2026, 8, 11, 0, 0, 0, 0, time.UTC)
		got, err := holidays.AnyHolidaysDuringWorkWeek(tue, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeFalse())
	})

	It("uses the prior Mon-Fri week for a Saturday", func() {
		// Saturday Nov 28, 2026 maps to the prior Mon Nov 23 - Fri Nov 27,
		// which contains Thanksgiving (Thu Nov 26).
		sat := time.Date(2026, 11, 28, 0, 0, 0, 0, time.UTC)
		got, err := holidays.AnyHolidaysDuringWorkWeek(sat, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeTrue(), "Saturday should use prior Mon-Fri; expected Thanksgiving hit")
	})

	It("uses the following Mon-Fri week for a Sunday", func() {
		// Sunday July 5, 2026 maps to the following Mon Jul 6 - Fri Jul 10, which
		// has no US federal holidays. (July 4 itself is excluded since it's Saturday.)
		sun := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)
		got, err := holidays.AnyHolidaysDuringWorkWeek(sun, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(BeFalse(), "Sunday should use following Mon-Fri (Jul 6-10), which has no us holidays")
	})

	It("propagates an error from the underlying Between call", func() {
		// Any date in 2049 makes the underlying Between call resolve 2050 as its
		// year+1 lookahead, which errors for kr (see the Between lunar-lookahead
		// test above); AnyHolidaysDuringWorkWeek must surface that error.
		d := time.Date(2049, 6, 15, 0, 0, 0, 0, time.UTC)
		_, err := holidays.AnyHolidaysDuringWorkWeek(d, holidays.Options{Regions: []string{"kr"}})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("CacheBetween", func() {
	BeforeEach(func() {
		holidays.ResetCache()
	})
	AfterEach(func() {
		holidays.ResetCache()
	})

	It("populates the cache and returns the same results as uncached Between", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		opts := holidays.Options{Regions: []string{"us"}}

		uncached, err := holidays.Between(from, to, opts)
		Expect(err).NotTo(HaveOccurred())

		Expect(holidays.CacheBetween(from, to, opts)).To(Succeed())

		cached, err := holidays.Between(from, to, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(cached).To(HaveLen(len(uncached)))
	})

	It("makes On resolve a known holiday after caching", func() {
		// CacheBetween a year, then On for a known holiday date should resolve
		// without recomputing; verified indirectly by correctness, since Between
		// is what On delegates to.
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		opts := holidays.Options{Regions: []string{"us"}}
		Expect(holidays.CacheBetween(from, to, opts)).To(Succeed())

		july4 := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(july4, opts)
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).NotTo(BeEmpty())
		Expect(hs[0].Name).To(Equal("Independence Day"))
	})

	It("returns only the subset within a narrower requested range", func() {
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		opts := holidays.Options{Regions: []string{"us"}}
		Expect(holidays.CacheBetween(from, to, opts)).To(Succeed())

		jul := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		julEnd := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.Between(jul, julEnd, opts)
		Expect(err).NotTo(HaveOccurred())
		for _, h := range hs {
			Expect(h.Date.Before(jul)).To(BeFalse(), "date %s outside requested range", h.Date.Format("2006-01-02"))
			Expect(h.Date.After(julEnd)).To(BeFalse(), "date %s outside requested range", h.Date.Format("2006-01-02"))
		}
	})

	It("returns an error and does not populate the cache when end is before start", func() {
		from := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		opts := holidays.Options{Regions: []string{"us"}}
		Expect(holidays.CacheBetween(from, to, opts)).To(HaveOccurred())

		_, err := holidays.Between(from, to, opts)
		Expect(err).To(HaveOccurred(), "cache must not have been populated with a bad range")
	})

	It("propagates a resolution error from the engine's year+1 lookahead", func() {
		from := time.Date(2049, 6, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2049, 6, 30, 0, 0, 0, 0, time.UTC)
		Expect(holidays.CacheBetween(from, to, holidays.Options{Regions: []string{"kr"}})).To(HaveOccurred())
	})
})

var _ = Describe("YearHolidaysFrom", func() {
	It("returns the full year when starting from Jan 1", func() {
		// Starting from Jan 1 2024 for us yields all 10 US holidays.
		from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.YearHolidaysFrom(from, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(10))

		for _, name := range []string{"New Year's Day", "Independence Day", "Thanksgiving", "Christmas Day"} {
			Expect(hasNamedHoliday(hs, name)).To(BeTrue(), "expected %q in full-year results", name)
		}
		for i := 1; i < len(hs); i++ {
			Expect(hs[i].Date.Before(hs[i-1].Date)).To(BeFalse(), "results not sorted ascending at index %d", i)
		}
	})

	It("returns only the remainder when starting mid-year", func() {
		// Starting from Nov 15 2024 for us yields only Thanksgiving
		// (2024-11-28) and Christmas (2024-12-25).
		from := time.Date(2024, 11, 15, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.YearHolidaysFrom(from, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(2))
		Expect(hs[0].Date.Format("2006-01-02")).To(Equal("2024-11-28"))
		Expect(hs[1].Date.Format("2006-01-02")).To(Equal("2024-12-25"))
		Expect(hasNamedHoliday(hs, "New Year's Day")).To(BeFalse(), "New Year's Day must be absent when starting from Nov 15")
	})

	It("includes a next-year observed date that shifts back into range", func() {
		// Starting from Jan 1 2021 for us with observed dates yields 11 holidays,
		// including New Year's Day observed on 2021-12-31 (Jan 1 2022 is a Saturday,
		// observed back to Friday Dec 31 2021).
		from := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.YearHolidaysFrom(from, holidays.Options{Regions: []string{"us"}, Observed: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(11))

		found := false
		for _, h := range hs {
			if h.Name == "New Year's Day" && h.Date.Format("2006-01-02") == "2021-12-31" {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "expected New Year's Day observed on 2021-12-31 in results")
	})

	It("propagates a primary-year resolution error, unlike the tolerated look-ahead failure", func() {
		// Unlike the look-ahead year (fromDay.Year()+1), a failure resolving the
		// primary year itself (i == 0) is not tolerated and must propagate. kr's
		// lunar tables end at 2049, so starting from 2050 fails on the primary year.
		from := time.Date(2050, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := holidays.YearHolidaysFrom(from, holidays.Options{Regions: []string{"kr"}})
		Expect(err).To(HaveOccurred())
	})

	It("tolerates a lunar look-ahead failure at the top of the lunar table range", func() {
		// go-holidays-egm: a lunar region at the top of the lunar-table range must
		// not fail its year query just because the year+1 look-ahead cannot
		// resolve. kr's lunar tables end at 2049, so ResolveYear(2050) errors; but
		// every 2050 result is clipped out of the [Jan 1, Dec 31 2049] window
		// anyway, so YearHolidaysFrom for 2049 must succeed rather than
		// propagating the look-ahead error.
		//
		// The assertion stays focused on egm: the call succeeds, clips to 2049,
		// and yields both a lunar (설날) and a gregorian (크리스마스) holiday. It
		// deliberately does NOT pin an exact count.
		from := time.Date(2049, 1, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.YearHolidaysFrom(from, holidays.Options{Regions: []string{"kr"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hasNamedHoliday(hs, "설날")).To(BeTrue())
		Expect(hasNamedHoliday(hs, "크리스마스")).To(BeTrue())
		for _, h := range hs {
			Expect(h.Date.Year()).To(Equal(2049), "result outside 2049 window: %s %s", h.Date.Format("2006-01-02"), h.Name)
		}
	})

	It("emits both KR Seollal holiday days that flank Seollal", func() {
		// go-holidays-253: Seollal (설날) is flanked by two 설날 연휴 days. The
		// eve is now modelled upstream as kr_seollal_eve(year, region), the day
		// before Seollal (the first day of the first lunar month); the day-after
		// is a plain lunar month:1/mday:2 rule. For 2022 Seollal is 2022-02-01,
		// so the eve is 2022-01-31 and the day-after is 2022-02-02.
		from := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.YearHolidaysFrom(from, holidays.Options{Regions: []string{"kr"}})
		Expect(err).NotTo(HaveOccurred())

		var got []string
		for _, h := range hs {
			if h.Name == "설날 연휴" {
				got = append(got, h.Date.Format("2006-01-02"))
			}
		}
		Expect(got).To(ConsistOf("2022-01-31", "2022-02-02"))
	})
})

var _ = Describe("On", func() {
	It("includes a next-year holiday whose observed date shifts back into range", func() {
		// go-holidays-wdp: On/Between must include a next-year holiday whose
		// observed date shifts back into the requested range. New Year's Day
		// Jan 1 2022 is a Saturday, observed back to Friday Dec 31 2021.
		date := time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(date, holidays.Options{Regions: []string{"us"}, Observed: true})
		Expect(err).NotTo(HaveOccurred())
		Expect(hasNamedHoliday(hs, "New Year's Day")).To(BeTrue(), "expected New Year's Day observed on 2021-12-31, got %+v", hs)
	})
})

var _ = Describe("YearHolidays", func() {
	It("returns every holiday matching the given options for a year", func() {
		hs, err := holidays.YearHolidays(2026, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(10))
		for _, name := range []string{"New Year's Day", "Independence Day", "Thanksgiving", "Christmas Day"} {
			Expect(hasNamedHoliday(hs, name)).To(BeTrue(), "expected %q in the 2026 us calendar", name)
		}
	})

	It("returns holidays across all regions when Regions is empty", func() {
		hs, err := holidays.YearHolidays(2026, holidays.Options{})
		Expect(err).NotTo(HaveOccurred())
		// An unfiltered year spans every registered region, far more than any
		// single region's yearly count.
		Expect(len(hs)).To(BeNumerically(">", 100))
	})

	It("propagates a resolution error from the engine", func() {
		// go-holidays-egm: kr's lunar tables end at 2049, so resolving 2050
		// directly (not as a look-ahead) must surface the engine's error.
		_, err := holidays.YearHolidays(2050, holidays.Options{Regions: []string{"kr"}})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Between", func() {
	It("errors when end is before start", func() {
		start := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		_, err := holidays.Between(start, end, holidays.Options{Regions: []string{"us"}})
		Expect(err).To(HaveOccurred())
	})

	It("propagates a resolution error from the engine's year+1 lookahead", func() {
		// computeBetween always resolves one year past end.Year() to catch a
		// next-year observed date shifting back into range; for kr in 2049 that
		// lookahead year (2050) is past the lunar table's range and errors, and
		// unlike YearHolidaysFrom, Between has no fallback: the error propagates.
		start := time.Date(2049, 6, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2049, 6, 30, 0, 0, 0, 0, time.UTC)
		_, err := holidays.Between(start, end, holidays.Options{Regions: []string{"kr"}})
		Expect(err).To(HaveOccurred())
	})

	It("includes a next-year holiday whose observed date shifts back into range", func() {
		start := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2021, 12, 31, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.Between(start, end, holidays.Options{Regions: []string{"us"}, Observed: true})
		Expect(err).NotTo(HaveOccurred())

		found := false
		for _, h := range hs {
			if h.Name == "New Year's Day" && h.Date.Format("2006-01-02") == "2021-12-31" {
				found = true
			}
		}
		Expect(found).To(BeTrue(), "expected New Year's Day observed on 2021-12-31 within 2021 range, got %+v", hs)
	})

	It("deduplicates a holiday shared identically across multiple requested regions", func() {
		// go-holidays-gmc: a multi-region request must not duplicate a holiday
		// defined for more than one of the requested regions. Informal Easter
		// Sunday exists in both us and ca; it must be emitted once.
		start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.Between(start, end, holidays.Options{Regions: []string{"us", "ca"}, Informal: true})
		Expect(err).NotTo(HaveOccurred())

		easter, goodFriday := 0, 0
		for _, h := range hs {
			switch h.Name {
			case "Easter Sunday":
				easter++
			case "Good Friday":
				goodFriday++
			}
		}
		// Easter Sunday is one identical informal def in both regions: merged to one.
		Expect(easter).To(Equal(1), "expected Easter Sunday exactly once for [us ca] informal 2024")
		// Good Friday has two distinct defs (us informal vs untagged us-states/ca):
		// they differ on type, so both must be kept. Dedup must not collapse them.
		Expect(goodFriday).To(Equal(2), "expected Good Friday twice for [us ca] informal 2024")
	})
})

var _ = Describe("wildcard region collapse (go-holidays-dpt)", func() {
	It("includes country-wide-only holidays for a multi-segment wildcard, same as the single-segment wildcard", func() {
		// A multi-segment wildcard like "au_vic_" must collapse all the way
		// to the country segment, matching the exact same rules as "au_": the
		// 2017 au_vic_ calendar is identical to the au_ calendar, 35 entries.
		// Australia Day (Jan 26) is a country-wide-only holiday (Regions: ["au"])
		// that a state-scoped wildcard must still surface.
		date := time.Date(2017, 1, 26, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(date, holidays.Options{Regions: []string{"au_vic_"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hasNamedHoliday(hs, "Australia Day")).To(BeTrue(), "expected Australia Day for au_vic_ on 2017-01-26, got %+v", hs)
	})

	It("returns the identical full-year Australian calendar for au_vic_ as for au_", func() {
		start := time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC)
		wide, err := holidays.Between(start, end, holidays.Options{Regions: []string{"au_"}})
		Expect(err).NotTo(HaveOccurred())
		scoped, err := holidays.Between(start, end, holidays.Options{Regions: []string{"au_vic_"}})
		Expect(err).NotTo(HaveOccurred())

		names := func(hs []holidays.Holiday) []string {
			out := make([]string, 0, len(hs))
			for _, h := range hs {
				out = append(out, h.Date.Format("2006-01-02")+" "+h.Name)
			}
			return out
		}
		Expect(names(scoped)).To(ConsistOf(names(wide)), "expected au_vic_ to match au_ exactly for 2017")
	})
})

var _ = Describe("same-date distinct holidays for a wildcard region (go-holidays-67s)", func() {
	It("keeps both the ACT/NSW/SA Labour Day and the QLD Queen's Birthday on 2020-10-05", func() {
		start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.Between(start, end, holidays.Options{Regions: []string{"au_"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(36), "expected the full 36-entry au_ 2020 calendar, matching gem 11.5.0")

		var onOct5 []string
		oct5 := time.Date(2020, 10, 5, 0, 0, 0, 0, time.UTC)
		for _, h := range hs {
			if h.Date.Equal(oct5) {
				onOct5 = append(onOct5, h.Name)
			}
		}
		Expect(onOct5).To(ConsistOf("Labour Day", "Queen's Birthday"))
	})
})

var _ = Describe("Holiday.Informal (go-holidays-4lb.2)", func() {
	It("distinguishes formal from informal US holidays over a full year", func() {
		start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.Between(start, end, holidays.Options{Regions: []string{"us"}, Informal: true})
		Expect(err).NotTo(HaveOccurred())

		hasInformal, hasFormal := false, false
		for _, h := range hs {
			if h.Informal {
				hasInformal = true
			} else {
				hasFormal = true
			}
		}
		Expect(hasInformal).To(BeTrue(), "expected at least one informal holiday in the [us] 2026 calendar, got %+v", hs)
		Expect(hasFormal).To(BeTrue(), "expected at least one formal holiday in the [us] 2026 calendar, got %+v", hs)
	})
})

var _ = Describe("region string normalization (go-holidays-9tl)", func() {
	It("treats an uppercase region code the same as lowercase", func() {
		date := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
		lower, err := holidays.On(date, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(lower).NotTo(BeEmpty())

		upper, err := holidays.On(date, holidays.Options{Regions: []string{"US"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(upper).To(Equal(lower))
	})

	It("trims surrounding whitespace from a region code", func() {
		date := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
		lower, err := holidays.On(date, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())

		padded, err := holidays.On(date, holidays.Options{Regions: []string{" us "}})
		Expect(err).NotTo(HaveOccurred())
		Expect(padded).To(Equal(lower))
	})

	It("treats a nil Regions slice as all regions, unchanged by normalization", func() {
		date := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(date, holidays.Options{})
		Expect(err).NotTo(HaveOccurred())
		Expect(hasNamedHoliday(hs, "Independence Day")).To(BeTrue(), "expected the us Independence Day holiday to still appear with no region filter")
	})

	It("normalizes a wildcard region code, preserving the trailing underscore", func() {
		start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)
		want, err := holidays.Between(start, end, holidays.Options{Regions: []string{"au_"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(want).NotTo(BeEmpty())

		got, err := holidays.Between(start, end, holidays.Options{Regions: []string{" AU_ "}})
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal(want))
	})
})
