package eventrouter_test

import (
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/caching/cachingfakes"
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/eventrouter"
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/logging/loggingfakes"
	. "github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events", func() {

	var r Router

	BeforeEach(func() {
		logging := new(FakeLogging)
		caching := new(FakeCaching)
		r = New(caching, logging)
		r.Setup("")

	})

	Context("called with a empty list", func() {
		It("should return a hash of only the default event", func() {
			expected := map[string]bool{"LogMessage": true}
			Expect(r.SelectedEvents()).To(Equal(expected))
		})
	})

	Context("called with a list of bogus event names", func() {
		It("should err out", func() {
			err := r.Setup("bogus,bogus1")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("called with a list of real event names", func() {
		It("should return a hash of events", func() {
			expected := map[string]bool{
				"HttpStartStop": true,
				"CounterEvent":  true,
			}
			r.Setup("HttpStartStop,CounterEvent")
			Expect(r.SelectedEvents()).To(Equal(expected))
		})
	})

	Context("called after 10 events have been routed", func() {
		var expected = uint64(10)
		BeforeEach(func() {
			for i := 0; i < int(expected); i++ {
				r.Route(&Envelope{EventType: Envelope_LogMessage.Enum()})
			}
		})

		It("should return a total of 10", func() {
			Expect(r.TotalCountOfSelectedEvents()).To(Equal(expected))
		})
	})

	Context("GetListAuthorizedEventEvents", func() {
		It("should return right list of authorized events", func() {
			Expect(GetListAuthorizedEventEvents()).To(Equal("ContainerMetric, CounterEvent, Error, HttpStart, HttpStartStop, HttpStop, LogMessage, ValueMetric"))
		})
	})

})
