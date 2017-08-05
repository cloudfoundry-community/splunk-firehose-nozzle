package eventsource_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEventsource(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Eventsource Suite")
}
