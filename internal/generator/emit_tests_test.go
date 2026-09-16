package generator

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("EmitTests", func() {
	Context("with a mix of expectation shapes and options", func() {
		It("emits a Describe block with one DescribeTable per test", func() {
			holidayTrue := true
			holidayFalse := false
			rf := &RegionFile{
				Country: "xx",
				Tests: []TestSpec{
					{
						Dates:    []string{"2024-1-1"},
						Regions:  []string{"xx"},
						Options:  []string{"informal", "observed"},
						Expected: ExpectedSpec{Name: "New Year's Day"},
					},
					{
						Dates:    []string{"2024-1-2"},
						Regions:  []string{"xx"},
						Expected: ExpectedSpec{Holiday: &holidayFalse},
					},
					{
						Dates:    []string{"2024-1-3"},
						Regions:  []string{"xx"},
						Expected: ExpectedSpec{Holiday: &holidayTrue},
					},
				},
			}

			src, err := EmitTests(rf)
			Expect(err).NotTo(HaveOccurred())
			out := string(src)

			Expect(out).To(ContainSubstring(`package definitions_test`))
			Expect(out).To(ContainSubstring(`var _ = Describe("xx", func() {`))
			Expect(out).To(ContainSubstring("Informal: true"))
			Expect(out).To(ContainSubstring("Observed: true"))
			Expect(out).To(ContainSubstring("hasNamed(hols"))
			Expect(out).To(ContainSubstring("Expect(hols).To(BeEmpty()"))
			Expect(out).To(ContainSubstring("Expect(hols).ToNot(BeEmpty()"))
		})
	})

	Context("when the country contains a newline that breaks out of the header comment", func() {
		It("returns the gofmt error alongside the raw bytes", func() {
			rf := &RegionFile{Country: "xx\n@@@not valid go@@@"}
			src, err := EmitTests(rf)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("gofmt"))
			Expect(src).NotTo(BeEmpty())
		})
	})
})

var _ = Describe("EmitTestHelpers", func() {
	It("returns the fixed bootstrap and shared-helper source", func() {
		out := string(EmitTestHelpers())
		Expect(out).To(ContainSubstring("package definitions_test"))
		Expect(out).To(ContainSubstring("func TestDefinitions(t *testing.T)"))
		Expect(out).To(ContainSubstring("func parseFlex(s string) (time.Time, error)"))
		Expect(out).To(ContainSubstring("func hasNamed(hs []holidays.Holiday, name string) bool"))
	})
})

var _ = Describe("writeDescribeTable", func() {
	It("emits one Entry per date and skips unrecognized options", func() {
		var buf bytes.Buffer
		writeDescribeTable(&buf, "xx", 0, TestSpec{
			Dates:    []string{"2024-1-1", "2024-1-2"},
			Regions:  []string{"xx"},
			Options:  []string{"informal", "something-else"},
			Expected: ExpectedSpec{Name: "N"},
		})
		out := buf.String()
		Expect(out).To(ContainSubstring(`Entry("2024-1-1", "2024-1-1")`))
		Expect(out).To(ContainSubstring(`Entry("2024-1-2", "2024-1-2")`))
		Expect(out).To(ContainSubstring("Informal: true"))
		Expect(out).NotTo(ContainSubstring("something-else"))
	})
})

var _ = Describe("testHint", func() {
	It("uses the sanitized expected name when one is given", func() {
		Expect(testHint(TestSpec{Expected: ExpectedSpec{Name: "New Year's Day"}})).To(Equal("NewYearSDay"))
	})

	It("returns Negative when holiday is explicitly false", func() {
		no := false
		Expect(testHint(TestSpec{Expected: ExpectedSpec{Holiday: &no}})).To(Equal("Negative"))
	})

	It("returns Positive when holiday is true or unspecified", func() {
		yes := true
		Expect(testHint(TestSpec{Expected: ExpectedSpec{Holiday: &yes}})).To(Equal("Positive"))
		Expect(testHint(TestSpec{})).To(Equal("Positive"))
	})
})

var _ = Describe("sanitizeIdent", func() {
	It("titlecases each word and strips punctuation/spaces", func() {
		Expect(sanitizeIdent("hello world")).To(Equal("HelloWorld"))
	})

	It("uppercases only the first rune of a run, leaving digits as-is", func() {
		Expect(sanitizeIdent("abc123 def")).To(Equal("Abc123Def"))
	})

	It("returns Unnamed for an empty string", func() {
		Expect(sanitizeIdent("")).To(Equal("Unnamed"))
	})

	It("returns Unnamed when nothing but punctuation is given", func() {
		Expect(sanitizeIdent("!!!")).To(Equal("Unnamed"))
	})
})
