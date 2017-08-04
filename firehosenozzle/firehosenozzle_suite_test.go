package firehosenozzle_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestFirehoseclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Firehoseclient Suite")
}
