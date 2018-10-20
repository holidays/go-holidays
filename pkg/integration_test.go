// +build integration

package holidays_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/holidays/go-holidays/pkg"
)

var _ = Describe("Integration Test - Between", func() {
	var (
		err     error
		regions []string
		options holidays.Options
	)

	BeforeEach(func() {
		err = nil
		regions = []string{"us"}
		options = holidays.Options{
			Informal: false,
			Observed: false,
		}
	})

	Describe("between", func() {
		var (
			start, end time.Time
			result     []holidays.Holiday
		)

		BeforeEach(func() {
			start = time.Date(2018, 1, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC)
		})

		JustBeforeEach(func() {
			result, err = holidays.Between(start, end, regions, options)
		})

		It("Does not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("Returns the expected two holidays for the us region in January 2018", func() {
			Expect(len(result)).To(Equal(2))
			Expect(result[0]).To(Equal(holidays.Holiday{
				Date:    start,
				Name:    "New Year's Day",
				Regions: []string{"us"},
			}))
			Expect(result[1]).To(Equal(holidays.Holiday{
				Date:    time.Date(2018, 1, 15, 0, 0, 0, 0, time.UTC),
				Name:    "Martin Luther King, Jr. Day",
				Regions: []string{"us"},
			}))
		})
	})

	Describe("available_regions", func() {
		var (
			regions []string
		)

		JustBeforeEach(func() {
			regions, err = holidays.AvailableRegions()
		})

		It("Does not return an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("Returns the correct number of regions", func() {
			Expect(len(regions)).To(Equal(240))
		})
	})
})
