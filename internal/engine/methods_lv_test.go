package engine

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lv_song_and_dance_festival_end_date", func() {
	call := func(year int) time.Time {
		fn, ok := LookupMethod("lv_song_and_dance_festival_end_date")
		Expect(ok).To(BeTrue())
		got, err := fn(MethodArgs{Year: year})
		Expect(err).NotTo(HaveOccurred())
		return got
	}

	It("returns the announced date for 2018", func() {
		Expect(call(2018)).To(Equal(time.Date(2018, time.July, 8, 0, 0, 0, 0, time.UTC)))
	})

	It("returns the announced date for 2023", func() {
		Expect(call(2023)).To(Equal(time.Date(2023, time.July, 9, 0, 0, 0, 0, time.UTC)))
	})

	It("returns zero for a year with no announced date", func() {
		Expect(call(2020).IsZero()).To(BeTrue())
	})
})
