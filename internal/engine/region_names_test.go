package engine_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/holidays/go-holidays/internal/engine"
)

var _ = Describe("RegisterRegionNames / RegionName / RegionNames", func() {
	It("registers names and returns them via comma-ok lookup", func() {
		engine.RegisterRegionNames("zz", map[string]string{
			"zz":     "Zzedonia",
			"zz_reg": "Zzedonia Region",
		})

		name, ok := engine.RegionName("zz")
		Expect(ok).To(BeTrue())
		Expect(name).To(Equal("Zzedonia"))

		name, ok = engine.RegionName("zz_reg")
		Expect(ok).To(BeTrue())
		Expect(name).To(Equal("Zzedonia Region"))
	})

	It("returns comma-ok false for an unregistered region", func() {
		_, ok := engine.RegionName("zz_does_not_exist")
		Expect(ok).To(BeFalse())
	})

	It("merges names from multiple RegisterRegionNames calls", func() {
		engine.RegisterRegionNames("zy", map[string]string{"zy": "Zyland"})
		engine.RegisterRegionNames("zx", map[string]string{"zx": "Zxland"})

		all := engine.RegionNames()
		Expect(all).To(HaveKeyWithValue("zy", "Zyland"))
		Expect(all).To(HaveKeyWithValue("zx", "Zxland"))
	})

	It("returns a copy that the caller cannot use to mutate internal state", func() {
		engine.RegisterRegionNames("zw", map[string]string{"zw": "Zwland"})

		all := engine.RegionNames()
		all["zw"] = "Mutated"

		name, ok := engine.RegionName("zw")
		Expect(ok).To(BeTrue())
		Expect(name).To(Equal("Zwland"))
	})
})
