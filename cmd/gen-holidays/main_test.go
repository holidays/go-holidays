package main

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// --- fixtures -----------------------------------------------------------

const versionTxt = "9.0.0\n"

// xxWithTests is a minimal, valid region YAML with a tests: block, so
// EmitTests and the country_test.go write path both run.
const xxWithTests = `
months:
  1:
    - name: "Xx Day"
      regions:
        - xx
      mday: 1
      type: formal

tests:
  - given:
      date: "2024-01-01"
      regions:
        - xx
    expect:
      name: "Xx Day"
      holiday: true
`

// yyNoTests is valid but has no tests: block, exercising the branch that
// removes any stale <country>_test.go instead of writing one.
const yyNoTests = `
months:
  1:
    - name: "Yy Day"
      regions:
        - yy
      mday: 2
      type: formal
`

// zzNoMonths has no months: block at all, so it parses to zero rules and
// must be skipped rather than generated.
const zzNoMonths = `
region_names:
  zz: "Zed Land"
`

// unportedFn references a function with no registered Go implementation.
const unportedFn = `
months:
  1:
    - name: "Unported Day"
      regions:
        - bad
      mday: 3
      type: formal
      function: totally_fake_method_xyz(1)
`

// malformedYAML fails strict decoding (unknown top-level field).
const malformedYAML = `
months:
  1:
    - name: "X"
      regions:
        - xx
      mday: 1
bogus_top_level_field: true
`

func writeFixture(dir string, files map[string]string) {
	GinkgoHelper()
	for name, content := range files {
		path := filepath.Join(dir, name)
		Expect(os.MkdirAll(filepath.Dir(path), 0o755)).To(Succeed())
		Expect(os.WriteFile(path, []byte(content), 0o644)).To(Succeed())
	}
}

// --- run() ---------------------------------------------------------------

var _ = Describe("run", func() {
	var (
		inDir  string
		outDir string
	)

	BeforeEach(func() {
		inDir = GinkgoT().TempDir()
		outDir = GinkgoT().TempDir()
	})

	Context("happy path", func() {
		It("generates regions, skips one with no months block, and handles the tests/no-tests split", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"xx.yaml":     xxWithTests,
				"yy.yaml":     yyNoTests,
				"zz.yaml":     zzNoMonths,
			})

			err := run([]string{"-in", inDir, "-out", outDir})
			Expect(err).NotTo(HaveOccurred())

			Expect(filepath.Join(outDir, "helpers_test.go")).To(BeAnExistingFile())
			Expect(filepath.Join(outDir, "xx.go")).To(BeAnExistingFile())
			Expect(filepath.Join(outDir, "xx_test.go")).To(BeAnExistingFile())
			Expect(filepath.Join(outDir, "yy.go")).To(BeAnExistingFile())
			Expect(filepath.Join(outDir, "yy_test.go")).NotTo(BeAnExistingFile())
			Expect(filepath.Join(outDir, "zz.go")).NotTo(BeAnExistingFile())
		})

		It("removes a stale test file for a region that no longer has tests", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"yy.yaml":     yyNoTests,
			})
			// Pre-seed a stale test file as if a previous run had tests.
			writeFixture(outDir, map[string]string{"yy_test.go": "package definitions_test\n"})

			Expect(run([]string{"-in", inDir, "-out", outDir})).To(Succeed())
			Expect(filepath.Join(outDir, "yy_test.go")).NotTo(BeAnExistingFile())
		})
	})

	Context("-regions filter", func() {
		It("generates only the requested region", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"xx.yaml":     xxWithTests,
				"yy.yaml":     yyNoTests,
			})

			Expect(run([]string{"-in", inDir, "-out", outDir, "-regions", "yy"})).To(Succeed())
			Expect(filepath.Join(outDir, "yy.go")).To(BeAnExistingFile())
			Expect(filepath.Join(outDir, "xx.go")).NotTo(BeAnExistingFile())
		})
	})

	Context("-allow-unported flag", func() {
		It("fails when a region references an unregistered method without the flag", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"bad.yaml":    unportedFn,
			})

			err := run([]string{"-in", inDir, "-out", outDir})
			Expect(err).To(MatchError(ContainSubstring("references unported methods")))
		})

		It("skips (rather than fails) the region when the flag is set", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"bad.yaml":    unportedFn,
			})

			Expect(run([]string{"-in", inDir, "-out", outDir, "-allow-unported"})).To(Succeed())
			Expect(filepath.Join(outDir, "bad.go")).NotTo(BeAnExistingFile())
		})
	})

	Context("errors", func() {
		It("rejects an unknown flag", func() {
			Expect(run([]string{"-nope"})).To(HaveOccurred())
		})

		It("fails when -in has no VERSION.txt", func() {
			err := run([]string{"-in", filepath.Join(inDir, "does-not-exist"), "-out", outDir})
			Expect(err).To(HaveOccurred())
		})

		It("fails when -regions names a region with no matching file", func() {
			writeFixture(inDir, map[string]string{"VERSION.txt": versionTxt})
			err := run([]string{"-in", inDir, "-out", outDir, "-regions", "nope"})
			Expect(err).To(MatchError(ContainSubstring("region nope")))
		})

		It("fails when no YAML files match", func() {
			writeFixture(inDir, map[string]string{"VERSION.txt": versionTxt})
			err := run([]string{"-in", inDir, "-out", outDir})
			Expect(err).To(MatchError("no input YAML files matched"))
		})

		It("fails when -out cannot be created", func() {
			blocker := filepath.Join(GinkgoT().TempDir(), "blocker")
			Expect(os.WriteFile(blocker, []byte("x"), 0o644)).To(Succeed())
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"xx.yaml":     xxWithTests,
			})

			err := run([]string{"-in", inDir, "-out", filepath.Join(blocker, "sub")})
			Expect(err).To(HaveOccurred())
		})

		It("fails when helpers_test.go cannot be written", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"xx.yaml":     xxWithTests,
			})
			Expect(os.Chmod(outDir, 0o555)).To(Succeed())
			DeferCleanup(func() { Expect(os.Chmod(outDir, 0o755)).To(Succeed()) })

			err := run([]string{"-in", inDir, "-out", outDir})
			Expect(err).To(MatchError(ContainSubstring("write helpers_test.go")))
		})

		It("fails when a selected input path is a directory instead of a file", func() {
			writeFixture(inDir, map[string]string{"VERSION.txt": versionTxt})
			Expect(os.MkdirAll(filepath.Join(inDir, "drx.yaml"), 0o755)).To(Succeed())

			err := run([]string{"-in", inDir, "-out", outDir, "-regions", "drx"})
			Expect(err).To(MatchError(ContainSubstring("read")))
		})

		It("fails when a region YAML fails strict parsing", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"xx.yaml":     malformedYAML,
			})

			err := run([]string{"-in", inDir, "-out", outDir})
			Expect(err).To(HaveOccurred())
		})

		It("fails when the country name is not a valid Go identifier, breaking gofmt", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"1x.yaml": `
months:
  1:
    - name: "One X Day"
      regions:
        - 1x
      mday: 1
      type: formal
`,
			})

			err := run([]string{"-in", inDir, "-out", outDir})
			Expect(err).To(HaveOccurred())
		})

		It("fails when the region's .go output path is already a directory", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"xx.yaml": `
months:
  1:
    - name: "Xx Day"
      regions:
        - xx
      mday: 1
      type: formal
`,
			})
			Expect(os.MkdirAll(outDir, 0o755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(outDir, "xx.go"), 0o755)).To(Succeed())

			err := run([]string{"-in", inDir, "-out", outDir})
			Expect(err).To(HaveOccurred())
		})

		It("fails when the region's _test.go output path is already a directory", func() {
			writeFixture(inDir, map[string]string{
				"VERSION.txt": versionTxt,
				"tt.yaml": `
months:
  1:
    - name: "Tt Day"
      regions:
        - tt
      mday: 1
      type: formal

tests:
  - given:
      date: "2024-01-01"
      regions:
        - tt
    expect:
      name: "Tt Day"
      holiday: true
`,
			})
			Expect(os.MkdirAll(outDir, 0o755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(outDir, "tt_test.go"), 0o755)).To(Succeed())

			err := run([]string{"-in", inDir, "-out", outDir})
			Expect(err).To(HaveOccurred())
		})
	})
})

// --- selectInputFiles ------------------------------------------------------

var _ = Describe("selectInputFiles", func() {
	var dir string

	BeforeEach(func() {
		dir = GinkgoT().TempDir()
	})

	It("returns all yaml files sorted, excluding index/METHODS and subdirectories", func() {
		writeFixture(dir, map[string]string{
			"zz.yaml":     zzNoMonths,
			"aa.yaml":     yyNoTests,
			"index.yaml":  "",
			"METHODS.yml": "",
			"notes.txt":   "",
		})
		Expect(os.MkdirAll(filepath.Join(dir, "subdir.yaml"), 0o755)).To(Succeed())

		got, err := selectInputFiles(dir, "")
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{
			filepath.Join(dir, "aa.yaml"),
			filepath.Join(dir, "zz.yaml"),
		}))
	})

	It("returns an error when the input directory does not exist", func() {
		_, err := selectInputFiles(filepath.Join(dir, "missing"), "")
		Expect(err).To(HaveOccurred())
	})

	It("filters to the requested regions, trimming whitespace and skipping blanks", func() {
		writeFixture(dir, map[string]string{
			"aa.yaml": yyNoTests,
			"zz.yaml": zzNoMonths,
		})

		got, err := selectInputFiles(dir, "aa, ,zz")
		Expect(err).NotTo(HaveOccurred())
		Expect(got).To(Equal([]string{
			filepath.Join(dir, "aa.yaml"),
			filepath.Join(dir, "zz.yaml"),
		}))
	})

	It("returns an error naming the missing region", func() {
		_, err := selectInputFiles(dir, "nope")
		Expect(err).To(MatchError(ContainSubstring("region nope")))
	})
})

// --- main() ----------------------------------------------------------------

var _ = Describe("main", func() {
	var (
		savedArgs []string
		savedExit func(int)
	)

	BeforeEach(func() {
		savedArgs = os.Args
		savedExit = osExit
	})

	AfterEach(func() {
		os.Args = savedArgs
		osExit = savedExit
	})

	It("runs to completion without exiting when the delegated run succeeds", func() {
		inDir := GinkgoT().TempDir()
		outDir := GinkgoT().TempDir()
		writeFixture(inDir, map[string]string{
			"VERSION.txt": versionTxt,
			"xx.yaml":     xxWithTests,
		})

		os.Args = []string{"gen-holidays", "-in", inDir, "-out", outDir}
		called := false
		osExit = func(int) { called = true }

		main()

		Expect(filepath.Join(outDir, "xx.go")).To(BeAnExistingFile())
		Expect(called).To(BeFalse())
	})

	It("prints the error and exits with status 1 when the run fails", func() {
		os.Args = []string{"gen-holidays", "-in", filepath.Join(GinkgoT().TempDir(), "does-not-exist")}
		var code int
		called := false
		osExit = func(c int) { called = true; code = c }

		main()

		Expect(called).To(BeTrue())
		Expect(code).To(Equal(1))
	})
})
