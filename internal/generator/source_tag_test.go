package generator

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("NormalizeSourceTag", func() {
	DescribeTable("normalizing the contents of VERSION.txt",
		func(raw, expected string) {
			tag, err := NormalizeSourceTag(raw)
			Expect(err).NotTo(HaveOccurred())
			Expect(tag).To(Equal(expected))
		},
		Entry("bare version gains a v prefix", "9.0.0", "v9.0.0"),
		Entry("prefixed version is left alone", "v9.0.0", "v9.0.0"),
		Entry("trailing newline is trimmed", "9.0.0\n", "v9.0.0"),
		Entry("surrounding whitespace is trimmed", "  \tv9.0.0 \r\n", "v9.0.0"),
	)

	Context("when the contents are blank", func() {
		It("returns an error rather than an empty tag", func() {
			_, err := NormalizeSourceTag(" \n\t")
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("LoadSourceTag", func() {
	Context("when VERSION.txt exists", func() {
		It("sets SourceTag from the normalized file contents", func() {
			dir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(dir, "VERSION.txt"), []byte("9.0.0"), 0o644)).To(Succeed())

			previous := SourceTag
			DeferCleanup(func() { SourceTag = previous })

			Expect(LoadSourceTag(dir)).To(Succeed())
			Expect(SourceTag).To(Equal("v9.0.0"))
		})
	})

	Context("when VERSION.txt is missing", func() {
		It("fails loudly and leaves SourceTag untouched", func() {
			previous := SourceTag
			DeferCleanup(func() { SourceTag = previous })

			err := LoadSourceTag(GinkgoT().TempDir())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("VERSION.txt"))
			Expect(SourceTag).To(Equal(previous))
		})
	})

	Context("when VERSION.txt is blank", func() {
		It("wraps the normalization error and leaves SourceTag untouched", func() {
			dir := GinkgoT().TempDir()
			Expect(os.WriteFile(filepath.Join(dir, "VERSION.txt"), []byte("  \n\t"), 0o644)).To(Succeed())

			previous := SourceTag
			DeferCleanup(func() { SourceTag = previous })

			err := LoadSourceTag(dir)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("VERSION.txt"))
			Expect(err.Error()).To(ContainSubstring("blank"))
			Expect(SourceTag).To(Equal(previous))
		})
	})
})
