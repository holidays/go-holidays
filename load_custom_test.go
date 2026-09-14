package holidays_test

import (
	"os"
	"path/filepath"
	"time"

	holidays "github.com/holidays/go-holidays"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func writeYAML(name, body string) string {
	path := filepath.Join(GinkgoT().TempDir(), name)
	Expect(os.WriteFile(path, []byte(body), 0o644)).To(Succeed())
	return path
}

var _ = Describe("LoadCustom", func() {
	It("makes a basic custom rule visible", func() {
		path := writeYAML("myteam.yaml", `
months:
  3:
  - name: Team Anniversary
    regions: [myteam]
    mday: 15
`)
		Expect(holidays.LoadCustom(path)).To(Succeed())
		defer holidays.UnloadCustom(path)

		d := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(d, holidays.Options{Regions: []string{"myteam"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(1))
		Expect(hs[0].Name).To(Equal("Team Anniversary"))
	})

	It("does not break built-in region lookups", func() {
		path := writeYAML("extras.yaml", `
months:
  6:
  - name: Custom June Holiday
    regions: [extras_test]
    mday: 10
`)
		Expect(holidays.LoadCustom(path)).To(Succeed())
		defer holidays.UnloadCustom(path)

		d := time.Date(2026, 7, 4, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(d, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).NotTo(BeEmpty())
		Expect(hs[0].Name).To(Equal("Independence Day"))
	})

	It("adds to an existing built-in region", func() {
		// A user can extend an existing region (e.g. "us") by writing a rule with
		// regions: [us]. The custom rule appears alongside the built-in ones.
		path := writeYAML("us_extras.yaml", `
months:
  4:
  - name: Custom April Holiday
    regions: [us]
    mday: 1
`)
		Expect(holidays.LoadCustom(path)).To(Succeed())
		defer holidays.UnloadCustom(path)

		d := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(d, holidays.Options{Regions: []string{"us"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hasNamedHoliday(hs, "Custom April Holiday")).To(BeTrue(), "expected Custom April Holiday in us results; got %+v", hs)
	})

	It("normalizes mixed-case and whitespace-padded custom region codes", func() {
		// The request path lowercases and trims Options.Regions, so a custom
		// rule's region codes must be normalized the same way at load time or
		// the region becomes unreachable by any query.
		path := writeYAML("MixedCase.yaml", `
months:
  6:
  - name: Mixed Case Region Day
    regions:
      - MyTeam
      - " Other_Team "
    mday: 15
`)
		Expect(holidays.LoadCustom(path)).To(Succeed())
		defer holidays.UnloadCustom(path)

		d := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
		for _, req := range []string{"MyTeam", "myteam", " MyTeam ", "MYTEAM"} {
			hs, err := holidays.On(d, holidays.Options{Regions: []string{req}})
			Expect(err).NotTo(HaveOccurred())
			Expect(hasNamedHoliday(hs, "Mixed Case Region Day")).To(BeTrue(),
				"request %q should resolve the custom region", req)
		}

		hs, err := holidays.On(d, holidays.Options{Regions: []string{"other_team"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hasNamedHoliday(hs, "Mixed Case Region Day")).To(BeTrue(),
			"whitespace-padded code should normalize to other_team")

		Expect(holidays.AvailableRegions()).To(ContainElement("myteam"))
		Expect(holidays.AvailableRegions()).NotTo(ContainElement("MyTeam"))
	})

	It("errors for a rule referencing an unregistered function", func() {
		path := writeYAML("bad.yaml", `
months:
  3:
  - name: Bad
    regions: [bad_test]
    function: not_a_real_method(year)
`)
		err := holidays.LoadCustom(path)
		if err == nil {
			holidays.UnloadCustom(path)
		}
		Expect(err).To(HaveOccurred())
	})

	It("resolves a rule via a method registered before loading", func() {
		// Register a custom method that always returns Feb 29 (in leap years).
		holidays.RegisterMethod("test_feb_29", func(a holidays.MethodArgs) (time.Time, error) {
			return time.Date(a.Year, time.February, 29, 0, 0, 0, 0, time.UTC), nil
		})

		path := writeYAML("leapday.yaml", `
months:
  2:
  - name: Leap Day
    regions: [leap_test]
    function: test_feb_29(year)
`)
		Expect(holidays.LoadCustom(path)).To(Succeed())
		defer holidays.UnloadCustom(path)

		d := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(d, holidays.Options{Regions: []string{"leap_test"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(1))
		Expect(hs[0].Name).To(Equal("Leap Day"))
	})

	It("errors when called with no paths", func() {
		Expect(holidays.LoadCustom()).To(HaveOccurred())
	})

	It("errors for a missing file", func() {
		Expect(holidays.LoadCustom("/nonexistent/path.yaml")).To(HaveOccurred())
	})

	It("replaces a previous load with the same basename on reload", func() {
		// Same basename, different contents: second load replaces first.
		dir := GinkgoT().TempDir()
		first := filepath.Join(dir, "swap.yaml")
		Expect(os.WriteFile(first, []byte(`
months:
  5:
  - name: First Version
    regions: [swap_test]
    mday: 1
`), 0o644)).To(Succeed())
		Expect(holidays.LoadCustom(first)).To(Succeed())

		Expect(os.WriteFile(first, []byte(`
months:
  5:
  - name: Second Version
    regions: [swap_test]
    mday: 1
`), 0o644)).To(Succeed())
		Expect(holidays.LoadCustom(first)).To(Succeed())
		defer holidays.UnloadCustom(first)

		d := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
		hs, err := holidays.On(d, holidays.Options{Regions: []string{"swap_test"}})
		Expect(err).NotTo(HaveOccurred())
		Expect(hs).To(HaveLen(1))
		Expect(hs[0].Name).To(Equal("Second Version"))
	})
})

var _ = Describe("UnloadCustom", func() {
	It("removes the rules it loaded", func() {
		path := writeYAML("temp.yaml", `
months:
  9:
  - name: Will Be Removed
    regions: [removable_test]
    mday: 5
`)
		Expect(holidays.LoadCustom(path)).To(Succeed())

		d := time.Date(2026, 9, 5, 0, 0, 0, 0, time.UTC)
		hs, _ := holidays.On(d, holidays.Options{Regions: []string{"removable_test"}})
		Expect(hs).To(HaveLen(1))

		holidays.UnloadCustom(path)

		hs, _ = holidays.On(d, holidays.Options{Regions: []string{"removable_test"}})
		Expect(hs).To(BeEmpty())
	})
})
