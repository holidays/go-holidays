package definition_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDefinition(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "definition suite")
}
