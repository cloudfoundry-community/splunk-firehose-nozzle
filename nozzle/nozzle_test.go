package nozzle_test

import (
	"time"

	"code.cloudfoundry.org/lager"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/testing"
)

var _ = Describe("Nozzle", func() {
	var (
		eventSource *testing.MemoryEventSourceMock
		eventRouter *testing.EventRouterMock
		nozzle      *Nozzle
	)

	Context("When there are no errors from event source", func() {
		BeforeEach(func() {
			eventSource = testing.NewMemoryEventSourceMock(-1, int64(10), -1)
			eventRouter = testing.NewEventRouterMock()
			config := &Config{
				Logger: lager.NewLogger("test"),
			}
			nozzle = New(eventSource, eventRouter, config)
		})

		It("collects events from source and routes to sink", func() {
			go nozzle.Start()

			time.Sleep(time.Second)

			Eventually(func() []*events.Envelope {
				return eventRouter.Events()
			}).Should(HaveLen(10))
		})

		It("EventSource close", func() {
			go nozzle.Start()
			time.Sleep(time.Second)
			eventSource.Close()
			time.Sleep(time.Second)
			nozzle.Close()
		})
	})

})
