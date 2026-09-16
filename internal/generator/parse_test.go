package generator

import (
	"gopkg.in/yaml.v3"

	"github.com/holidays/go-holidays/internal/definition"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Verifies that a YAML containing a top-level region_names: block (added in
// v8.0.0 of the upstream definitions) does not cause ParseRegionFile to error
// when strict decoding is enabled.
var _ = Describe("ParseRegionFile", func() {
	Context("when the YAML has a top-level region_names block", func() {
		It("parses without error and keeps the rule intact", func() {
			yaml := []byte(`
months:
  1:
    - name: "New Year's Day"
      regions:
        - xx
      mday: 1
      type: formal

region_names:
  xx: "Example Country"
  xx_reg: "Example Region"

tests:
  - given:
      date: "2024-01-01"
      regions:
        - xx
    expect:
      name: "New Year's Day"
      holiday: true
`)

			rf, err := ParseRegionFile("xx", yaml)
			Expect(err).NotTo(HaveOccurred())
			Expect(rf.Rules).To(HaveLen(1))
			Expect(rf.Rules[0].Name).To(Equal("New Year's Day"))
			Expect(rf.RegionNames).To(Equal(map[string]string{
				"xx":     "Example Country",
				"xx_reg": "Example Region",
			}))
		})
	})

	Context("when the YAML has no region_names block", func() {
		It("leaves RegionNames nil", func() {
			yaml := []byte(`
months:
  1:
    - name: "New Year's Day"
      regions:
        - xx
      mday: 1
      type: formal
`)

			rf, err := ParseRegionFile("xx", yaml)
			Expect(err).NotTo(HaveOccurred())
			Expect(rf.RegionNames).To(BeEmpty())
		})
	})

	Context("when the YAML is malformed", func() {
		It("wraps the decode error", func() {
			_, err := ParseRegionFile("xx", []byte("months: [this is not a mapping"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("parse xx.yaml"))
		})
	})

	Context("when a month entry fails to convert", func() {
		It("wraps the error with the month and entry index", func() {
			bad := []byte(`
months:
  1:
    - regions:
        - xx
      mday: 1
`)
			_, err := ParseRegionFile("xx", bad)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("xx.yaml month 1 entry 0"))
			Expect(err.Error()).To(ContainSubstring("missing name"))
		})
	})

	Context("when a test entry fails to convert", func() {
		It("wraps the error with the test index", func() {
			bad := []byte(`
months:
  1:
    - name: "New Year's Day"
      regions:
        - xx
      mday: 1

tests:
  - given:
      date: "2024-01-01"
      regions:
        - xx
    expect: {}
`)
			_, err := ParseRegionFile("xx", bad)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("xx.yaml test 0"))
		})
	})

	Context("when a test entry has only unparseable dates", func() {
		It("drops the test entirely rather than erroring", func() {
			src := []byte(`
months:
  1:
    - name: "New Year's Day"
      regions:
        - xx
      mday: 1

tests:
  - given:
      date: "not-a-real-date"
      regions:
        - xx
    expect:
      name: "New Year's Day"
`)
			rf, err := ParseRegionFile("xx", src)
			Expect(err).NotTo(HaveOccurred())
			Expect(rf.Tests).To(BeEmpty())
		})
	})

	Context("when methods: is present", func() {
		It("captures name, split arguments, and the ruby body", func() {
			src := []byte(`
months:
  1:
    - name: "New Year's Day"
      regions:
        - xx
      mday: 1

methods:
  custom_thing:
    arguments: "year, month"
    ruby: "def custom_thing; end"
`)
			rf, err := ParseRegionFile("xx", src)
			Expect(err).NotTo(HaveOccurred())
			Expect(rf.Methods).To(HaveKey("custom_thing"))
			spec := rf.Methods["custom_thing"]
			Expect(spec.Name).To(Equal("custom_thing"))
			Expect(spec.Arguments).To(Equal([]string{"year", "month"}))
			Expect(spec.Ruby).To(Equal("def custom_thing; end"))
		})
	})
})

var _ = Describe("convertRule", func() {
	It("errors when name is missing", func() {
		_, err := convertRule("xx", 1, rawRule{Regions: []string{"xx"}})
		Expect(err).To(MatchError("missing name"))
	})

	It("errors when regions is empty", func() {
		_, err := convertRule("xx", 1, rawRule{Name: "N"})
		Expect(err).To(MatchError("missing regions"))
	})

	It("uses the explicit wday when given", func() {
		wday := 3
		rule, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}, Wday: &wday})
		Expect(err).NotTo(HaveOccurred())
		Expect(rule.Wday).To(Equal(3))
	})

	It("defaults wday to -1 when not given", func() {
		rule, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(rule.Wday).To(Equal(-1))
	})

	It("treats an empty type as formal", func() {
		rule, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(rule.Type).To(Equal(definition.Formal))
	})

	It("accepts informal (case-insensitively)", func() {
		rule, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}, Type: "INFORMAL"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rule.Type).To(Equal(definition.Informal))
	})

	It("errors on an unknown type", func() {
		_, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}, Type: "bogus"})
		Expect(err).To(MatchError(`unknown type "bogus"`))
	})

	It("parses a function call and its arguments", func() {
		rule, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}, Function: "foo(a, b)"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rule.Function).To(Equal("foo"))
		Expect(rule.FunctionArgs).To(Equal([]string{"a", "b"}))
	})

	It("wraps an invalid function expression", func() {
		_, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}, Function: "not valid("})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`function "not valid("`))
	})

	It("parses an observed call", func() {
		rule, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}, Observed: "to_monday_if_sunday()"})
		Expect(err).NotTo(HaveOccurred())
		Expect(rule.Observed).To(Equal("to_monday_if_sunday"))
	})

	It("wraps an invalid observed expression", func() {
		_, err := convertRule("xx", 1, rawRule{Name: "N", Regions: []string{"xx"}, Observed: "not valid("})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring(`observed "not valid("`))
	})

	It("converts a valid year_ranges block", func() {
		until := 2020
		rule, err := convertRule("xx", 1, rawRule{
			Name: "N", Regions: []string{"xx"},
			YearRanges: &rawYearRanges{Until: &until},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(rule.YearRanges).To(HaveLen(1))
		Expect(rule.YearRanges[0].Kind).To(Equal(definition.YearRangeUntil))
	})

	It("propagates a year_ranges conversion error", func() {
		_, err := convertRule("xx", 1, rawRule{
			Name: "N", Regions: []string{"xx"},
			YearRanges: &rawYearRanges{},
		})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("convertYearRange", func() {
	It("errors when no field is set", func() {
		_, err := convertYearRange(rawYearRanges{})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("got 0"))
	})

	It("errors when more than one field is set", func() {
		until, from := 2020, 1990
		_, err := convertYearRange(rawYearRanges{Until: &until, From: &from})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("got 2"))
	})

	It("converts until", func() {
		until := 2020
		yr, err := convertYearRange(rawYearRanges{Until: &until})
		Expect(err).NotTo(HaveOccurred())
		Expect(yr).To(Equal(definition.YearRange{Kind: definition.YearRangeUntil, Years: []int{2020}}))
	})

	It("converts from", func() {
		from := 1990
		yr, err := convertYearRange(rawYearRanges{From: &from})
		Expect(err).NotTo(HaveOccurred())
		Expect(yr).To(Equal(definition.YearRange{Kind: definition.YearRangeFrom, Years: []int{1990}}))
	})

	It("converts limited", func() {
		yr, err := convertYearRange(rawYearRanges{Limited: []int{1, 2, 3}})
		Expect(err).NotTo(HaveOccurred())
		Expect(yr).To(Equal(definition.YearRange{Kind: definition.YearRangeLimited, Years: []int{1, 2, 3}}))
	})

	It("converts between", func() {
		yr, err := convertYearRange(rawYearRanges{Between: &rawBetweenSpan{Start: 1, End: 2}})
		Expect(err).NotTo(HaveOccurred())
		Expect(yr).To(Equal(definition.YearRange{Kind: definition.YearRangeBetween, Years: []int{1, 2}}))
	})
})

var _ = Describe("convertTest", func() {
	It("errors when given.date fails to decode", func() {
		_, _, err := convertTest("xx", 0, rawTest{
			Given: rawTestGiven{Date: yaml.Node{Kind: yaml.MappingNode}},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("given.date"))
	})

	It("wraps the missing-date error from an empty given.date", func() {
		_, _, err := convertTest("xx", 0, rawTest{
			Given: rawTestGiven{Date: yaml.Node{Kind: yaml.ScalarNode, Value: ""}},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("given.date"))
		Expect(err.Error()).To(ContainSubstring("missing date"))
	})

	It("skips a test whose only date is unparseable", func() {
		_, skip, err := convertTest("xx", 0, rawTest{
			Given:  rawTestGiven{Date: yaml.Node{Kind: yaml.ScalarNode, Value: "not-a-date"}},
			Expect: rawTestExpect{Name: "N"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(skip).To(BeTrue())
	})

	It("drops unparseable dates but keeps the rest", func() {
		ts, skip, err := convertTest("xx", 0, rawTest{
			Given: rawTestGiven{Date: yaml.Node{
				Kind: yaml.SequenceNode,
				Content: []*yaml.Node{
					{Kind: yaml.ScalarNode, Value: "2024-1-1"},
					{Kind: yaml.ScalarNode, Value: "not-a-date"},
				},
			}},
			Expect: rawTestExpect{Name: "N"},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(skip).To(BeFalse())
		Expect(ts.Dates).To(Equal([]string{"2024-1-1"}))
	})

	It("errors when given.options fails to decode", func() {
		_, _, err := convertTest("xx", 0, rawTest{
			Given:  rawTestGiven{Date: yaml.Node{Kind: yaml.ScalarNode, Value: "2024-1-1"}, Options: yaml.Node{Kind: yaml.MappingNode}},
			Expect: rawTestExpect{Name: "N"},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("given.options"))
	})

	It("errors when expect specifies neither name nor holiday", func() {
		_, _, err := convertTest("xx", 0, rawTest{
			Given: rawTestGiven{Date: yaml.Node{Kind: yaml.ScalarNode, Value: "2024-1-1"}},
		})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("expect must specify name or holiday"))
	})

	It("falls back to the holyday alias when holiday is absent", func() {
		want := true
		ts, _, err := convertTest("xx", 0, rawTest{
			Given:  rawTestGiven{Date: yaml.Node{Kind: yaml.ScalarNode, Value: "2024-1-1"}},
			Expect: rawTestExpect{Holyday: &want},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(ts.Expected.Holiday).To(Equal(&want))
	})
})

var _ = Describe("filterParseableDates", func() {
	It("splits valid and invalid dates", func() {
		valid, skipped := filterParseableDates([]string{"2024-1-1", "nope", "2024-12-31"})
		Expect(valid).To(Equal([]string{"2024-1-1", "2024-12-31"}))
		Expect(skipped).To(Equal([]string{"nope"}))
	})
})

var _ = Describe("decodeDateNode", func() {
	It("propagates a decode error", func() {
		_, err := decodeDateNode(yaml.Node{Kind: yaml.MappingNode})
		Expect(err).To(HaveOccurred())
	})

	It("errors when the result is empty", func() {
		_, err := decodeDateNode(yaml.Node{Kind: 0})
		Expect(err).To(MatchError("missing date"))
	})

	It("returns the decoded list", func() {
		out, err := decodeDateNode(yaml.Node{Kind: yaml.ScalarNode, Value: "2024-1-1"})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal([]string{"2024-1-1"}))
	})
})

var _ = Describe("decodeStringOrList", func() {
	It("returns nil for an unset node", func() {
		out, err := decodeStringOrList(yaml.Node{Kind: 0})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeNil())
	})

	It("returns nil for an empty scalar", func() {
		out, err := decodeStringOrList(yaml.Node{Kind: yaml.ScalarNode, Value: ""})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(BeNil())
	})

	It("wraps a non-empty scalar in a single-element slice", func() {
		out, err := decodeStringOrList(yaml.Node{Kind: yaml.ScalarNode, Value: "hi"})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal([]string{"hi"}))
	})

	It("decodes a sequence of scalars", func() {
		out, err := decodeStringOrList(yaml.Node{
			Kind: yaml.SequenceNode,
			Content: []*yaml.Node{
				{Kind: yaml.ScalarNode, Value: "a"},
				{Kind: yaml.ScalarNode, Value: "b"},
			},
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal([]string{"a", "b"}))
	})

	It("errors when a sequence contains a non-scalar", func() {
		_, err := decodeStringOrList(yaml.Node{
			Kind:    yaml.SequenceNode,
			Content: []*yaml.Node{{Kind: yaml.MappingNode}},
		})
		Expect(err).To(MatchError("list contains non-scalar"))
	})

	It("errors on an unexpected node kind", func() {
		_, err := decodeStringOrList(yaml.Node{Kind: yaml.MappingNode})
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("unexpected node kind"))
	})
})

var _ = Describe("splitArgs", func() {
	It("returns nil for a blank string", func() {
		Expect(splitArgs("   ")).To(BeNil())
	})

	It("trims and splits comma-separated arguments", func() {
		Expect(splitArgs("a, b ,c")).To(Equal([]string{"a", "b", "c"}))
	})
})

var _ = Describe("ParseCall", func() {
	It("errors on an expression that doesn't match name(args)", func() {
		_, err := ParseCall("not valid(")
		Expect(err).To(HaveOccurred())
	})

	It("parses a call with no arguments", func() {
		mc, err := ParseCall("foo()")
		Expect(err).NotTo(HaveOccurred())
		Expect(mc).To(Equal(MethodCall{Name: "foo"}))
	})

	It("parses a call with arguments", func() {
		mc, err := ParseCall("foo(a, b)")
		Expect(err).NotTo(HaveOccurred())
		Expect(mc.Name).To(Equal("foo"))
		Expect(mc.Args).To(Equal([]string{"a", "b"}))
	})
})
