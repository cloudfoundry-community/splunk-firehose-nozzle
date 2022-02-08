package eventfilter_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEventfilter(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Eventfilter Suite")
}
