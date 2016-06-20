package nozzle_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestNozzle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Nozzle Suite")
}
