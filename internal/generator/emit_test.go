package generator

import (
	"bytes"

	"github.com/holidays/go-holidays/internal/definition"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("EmitDefinitions region names", func() {
	Context("when the region file has region_names", func() {
		It("emits a sorted map var and registers it in init()", func() {
			rf := &RegionFile{
				Country: "xx",
				RegionNames: map[string]string{
					"xx_reg": "Example Region",
					"xx":     "Example Country",
				},
			}

			src, err := EmitDefinitions(rf)
			Expect(err).NotTo(HaveOccurred())
			out := string(src)

			Expect(out).To(ContainSubstring(`var xxRegionNames = map[string]string{`))
			// sorted by key: "xx" before "xx_reg"
			xxIdx := indexOf(out, `"Example Country"`)
			xxRegIdx := indexOf(out, `"Example Region"`)
			Expect(xxIdx).To(BeNumerically(">=", 0))
			Expect(xxRegIdx).To(BeNumerically(">", xxIdx))

			Expect(out).To(ContainSubstring(`engine.RegisterCountry("xx", xxRules)`))
			Expect(out).To(ContainSubstring(`engine.RegisterRegionNames("xx", xxRegionNames)`))
		})
	})

	Context("when the region file has no region_names", func() {
		It("does not emit a RegionNames var or RegisterRegionNames call", func() {
			rf := &RegionFile{Country: "xx"}

			src, err := EmitDefinitions(rf)
			Expect(err).NotTo(HaveOccurred())
			out := string(src)

			Expect(out).NotTo(ContainSubstring("RegionNames"))
			Expect(out).NotTo(ContainSubstring("RegisterRegionNames"))
		})
	})
})

var _ = Describe("ValidationError", func() {
	It("formats the country and the sorted missing methods", func() {
		err := &ValidationError{Country: "xx", Missing: []string{"foo", "bar"}}
		Expect(err.Error()).To(Equal("xx.yaml references unported methods: foo, bar"))
	})
})

var _ = Describe("Validate", func() {
	Context("when every referenced function/observed method is registered", func() {
		It("returns nil", func() {
			rf := &RegionFile{Country: "xx", Rules: []definition.HolidayRule{
				{Name: "N", Function: "easter", Observed: "to_monday_if_sunday"},
				{Name: "M"},
			}}
			Expect(Validate(rf)).To(Succeed())
		})
	})

	Context("when function/observed references have no registered implementation", func() {
		It("returns a sorted, de-duplicated ValidationError", func() {
			rf := &RegionFile{Country: "xx", Rules: []definition.HolidayRule{
				{Name: "N", Function: "zzz_missing"},
				{Name: "M", Observed: "aaa_missing"},
				{Name: "O", Function: "zzz_missing"},
			}}
			err := Validate(rf)
			Expect(err).To(HaveOccurred())
			ve, ok := err.(*ValidationError)
			Expect(ok).To(BeTrue())
			Expect(ve.Country).To(Equal("xx"))
			Expect(ve.Missing).To(Equal([]string{"aaa_missing", "zzz_missing"}))
		})
	})
})

var _ = Describe("EmitDefinitions", func() {
	Context("when the region file has rules", func() {
		It("emits one rule literal per rule, in order", func() {
			rf := &RegionFile{
				Country: "xx",
				Rules: []definition.HolidayRule{
					{Name: "New Year's Day", Regions: []string{"xx"}, Month: 1, Mday: 1, Wday: -1},
					{Name: "Some Other Day", Regions: []string{"xx"}, Month: 5, Wday: -1},
				},
			}
			src, err := EmitDefinitions(rf)
			Expect(err).NotTo(HaveOccurred())
			out := string(src)
			Expect(out).To(ContainSubstring(`"New Year's Day"`))
			Expect(out).To(ContainSubstring(`"Some Other Day"`))
			Expect(out).To(ContainSubstring("var xxRules = []definition.HolidayRule{"))
			Expect(out).To(ContainSubstring(`engine.RegisterCountry("xx", xxRules)`))
		})
	})

	Context("when the country produces invalid Go source", func() {
		It("returns the gofmt error alongside the raw bytes", func() {
			rf := &RegionFile{Country: "1bad"}
			src, err := EmitDefinitions(rf)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("gofmt 1bad"))
			Expect(src).NotTo(BeEmpty())
		})
	})
})

var _ = Describe("writeRuleLiteral", func() {
	It("omits zero-value fields and prints wday -1 when unset", func() {
		var buf bytes.Buffer
		writeRuleLiteral(&buf, definition.HolidayRule{Name: "A", Regions: []string{"xx"}, Month: 1, Wday: -1})
		out := buf.String()
		Expect(out).To(ContainSubstring(`Name: "A"`))
		Expect(out).To(ContainSubstring("Wday: -1,"))
		Expect(out).NotTo(ContainSubstring("Mday:"))
		Expect(out).NotTo(ContainSubstring("Week:"))
		Expect(out).NotTo(ContainSubstring("Function:"))
		Expect(out).NotTo(ContainSubstring("FunctionModifier:"))
		Expect(out).NotTo(ContainSubstring("Observed:"))
		Expect(out).NotTo(ContainSubstring("definition.Informal"))
		Expect(out).NotTo(ContainSubstring("YearRanges:"))
	})

	It("prints a function with no arguments without a FunctionArgs field", func() {
		var buf bytes.Buffer
		writeRuleLiteral(&buf, definition.HolidayRule{Name: "B", Regions: []string{"xx"}, Wday: -1, Function: "f3"})
		out := buf.String()
		Expect(out).To(ContainSubstring(`Function: "f3"`))
		Expect(out).NotTo(ContainSubstring("FunctionArgs:"))
	})

	It("prints every optional field and every year-range kind, including the unknown default", func() {
		var buf bytes.Buffer
		writeRuleLiteral(&buf, definition.HolidayRule{
			Name:             "C",
			Regions:          []string{"xx", "yy"},
			Month:            2,
			Mday:             15,
			Wday:             3,
			Week:             2,
			Function:         "f1",
			FunctionArgs:     []string{"x"},
			FunctionModifier: 1,
			Observed:         "o1",
			Type:             definition.Informal,
			YearRanges: []definition.YearRange{
				{Kind: definition.YearRangeUntil, Years: []int{2020}},
				{Kind: definition.YearRangeFrom, Years: []int{1990}},
				{Kind: definition.YearRangeLimited, Years: []int{1, 2, 3}},
				{Kind: definition.YearRangeBetween, Years: []int{1, 2}},
				{Kind: definition.YearRangeAny, Years: nil},
			},
		})
		out := buf.String()
		Expect(out).To(ContainSubstring("Mday: 15,"))
		Expect(out).To(ContainSubstring("Wday: 3,"))
		Expect(out).To(ContainSubstring("Week: 2,"))
		Expect(out).To(ContainSubstring(`Function: "f1"`))
		Expect(out).To(ContainSubstring(`FunctionArgs: []string{"x"}`))
		Expect(out).To(ContainSubstring("FunctionModifier: 1,"))
		Expect(out).To(ContainSubstring(`Observed: "o1"`))
		Expect(out).To(ContainSubstring("definition.Informal"))
		Expect(out).To(ContainSubstring("definition.YearRangeUntil"))
		Expect(out).To(ContainSubstring("definition.YearRangeFrom"))
		Expect(out).To(ContainSubstring("definition.YearRangeLimited"))
		Expect(out).To(ContainSubstring("definition.YearRangeBetween"))
		Expect(out).To(ContainSubstring("definition.YearRangeAny"))
	})
})

var _ = Describe("yearRangeKindName", func() {
	DescribeTable("mapping YearRangeKind to its Go source identifier",
		func(kind definition.YearRangeKind, want string) {
			Expect(yearRangeKindName(kind)).To(Equal(want))
		},
		Entry("until", definition.YearRangeUntil, "definition.YearRangeUntil"),
		Entry("from", definition.YearRangeFrom, "definition.YearRangeFrom"),
		Entry("limited", definition.YearRangeLimited, "definition.YearRangeLimited"),
		Entry("between", definition.YearRangeBetween, "definition.YearRangeBetween"),
		Entry("any (default)", definition.YearRangeAny, "definition.YearRangeAny"),
	)
})

var _ = Describe("stringSliceLit", func() {
	It("returns nil for an empty slice", func() {
		Expect(stringSliceLit(nil)).To(Equal("nil"))
	})

	It("renders a quoted, comma-joined literal", func() {
		Expect(stringSliceLit([]string{"a", "b"})).To(Equal(`[]string{"a", "b"}`))
	})
})

var _ = Describe("intSliceLit", func() {
	It("returns nil for an empty slice", func() {
		Expect(intSliceLit(nil)).To(Equal("nil"))
	})

	It("renders a comma-joined literal", func() {
		Expect(intSliceLit([]int{1, 2})).To(Equal("[]int{1, 2}"))
	})
})

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}
