package holidays_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHolidays(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Main Holidays Suite")
}
