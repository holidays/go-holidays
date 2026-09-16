package calc

import "time"

// hijriEpochJDN is the Julian day number of 1 Muharram 1 AH minus one day,
// chosen so that hijriToJDN's month/day terms are added directly. It
// corresponds to the civil epoch (Friday, 16 July 622 CE proleptic Julian).
// Ported from Holidays::DateCalculator::HijriDate::EPOCH_JDN in the Ruby
// holidays gem.
const hijriEpochJDN = 1948439

// unixEpochJDN is the Julian day number of 1970-01-01 (Gregorian), used to
// translate a Julian day number into a time.Time via the standard library's
// calendar arithmetic instead of a bespoke JDN-to-Gregorian algorithm.
const unixEpochJDN = 2440588

var unixEpoch = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)

// HijriToGregorian converts a date in the arithmetic ("tabular", a.k.a.
// Kuwaiti) Islamic calendar to its Gregorian equivalent.
//
// Months alternate 30 and 29 days; leap years add a day to the twelfth month
// on a fixed 11-of-30 cycle. This is a purely arithmetic calendar, so it can
// land a day (occasionally two) either side of a locally proclaimed date.
// Callers that need the proclaimed date layer their own override table on
// top of this (see engine.trFeastDate for tr Ramazan/Kurban Bayramı).
func HijriToGregorian(hijriYear, hijriMonth, hijriDay int) time.Time {
	return unixEpoch.AddDate(0, 0, hijriToJDN(hijriYear, hijriMonth, hijriDay)-unixEpochJDN)
}

// HijriYearOccurrence returns the Gregorian occurrence of a fixed Hijri
// month/day that falls within gregorianYear, and true if one was found.
//
// A fixed Hijri date drifts about 11 days earlier each Gregorian year, so
// roughly once every 33 years it falls twice in the same Gregorian year
// (once in early January, once in late December). Only the earlier
// occurrence is returned.
func HijriYearOccurrence(gregorianYear, hijriMonth, hijriDay int) (time.Time, bool) {
	candidateHijriYear := gregorianYear - 579
	for _, hijriYear := range [3]int{candidateHijriYear - 1, candidateHijriYear, candidateHijriYear + 1} {
		date := HijriToGregorian(hijriYear, hijriMonth, hijriDay)
		if date.Year() == gregorianYear {
			return date, true
		}
	}
	return time.Time{}, false
}

// hijriToJDN computes the Julian day number for a Hijri calendar date.
func hijriToJDN(year, month, day int) int {
	return day + ceilHalf(59*(month-1)) + (year-1)*354 + (3+11*year)/30 + hijriEpochJDN
}

// ceilHalf returns ceil(n / 2.0) for a non-negative integer n.
func ceilHalf(n int) int {
	return (n + 1) / 2
}
