package utils_test

import (
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Utils counter test", func() {

	BeforeEach(func() {})

	It("Add counters", func() {
		ctr := new(utils.IntCounter)
		ctr.Add(uint64(5))
		ctr2 := new(utils.IntCounter)
		ctr2.Add(uint64(3))
		ctr2.Add(*ctr)
		Expect(ctr2.Value()).To(Equal(uint64(8)))
	})

	It("Clone counter", func() {
		ctr := new(utils.IntCounter)
		ctr.Add(8)
		ctr1 := ctr.Clone()
		Expect(ctr1.Value()).To(Equal(uint64(8)))
	})

	It("Clone value when original pointer becomes zero", func() {
		ctr := new(utils.IntCounter)
		ctr.Add(8)
		ctr1 := ctr.Clone()
		ctr.Reset()
		Expect(ctr1.Value()).To(Equal(uint64(8)))
	})
})
