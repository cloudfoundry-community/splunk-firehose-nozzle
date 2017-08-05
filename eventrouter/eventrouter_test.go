package eventrouter_test

import (
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/cache/cachefakes"
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/eventrouter"
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsink/eventsinkfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("eventrouter", func() {

	var r Router
	var err error

	BeforeEach(func() {
		sink := new(FakeSink)
		caching := new(FakeCache)
		config := &Config{
			SelectedEvents: "",
		}
		r, err = New(caching, sink, config)
		Ω(err).ShouldNot(HaveOccurred())
	})

})
