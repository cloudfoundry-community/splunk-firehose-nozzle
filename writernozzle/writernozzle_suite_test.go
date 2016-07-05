package writernozzle_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWriternozzle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Writernozzle Suite")
}
