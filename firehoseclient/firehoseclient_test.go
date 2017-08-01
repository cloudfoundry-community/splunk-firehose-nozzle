package firehoseclient_test

import (
	"time"

	"code.cloudfoundry.org/lager"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/firehoseclient"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/testing"
)

var _ = Describe("Firehoseclient", func() {
	var (
		consumer    *testing.MockFirehoseConsumer
		eventRouter *testing.MockEventRouter
		nozzle      *FirehoseNozzle
	)

	Context("When there are no errors from event source", func() {
		BeforeEach(func() {
			consumer = testing.NewMockFirehoseConsumer(-1, int64(10), -1)
			eventRouter = testing.NewMockEventRouter()
			config := &FirehoseConfig{
				FirehoseSubscriptionID: "splunk-subcription-id",

				Logger: lager.NewLogger("test"),
			}
			nozzle = NewFirehoseNozzle(consumer, eventRouter, config)
		})

		It("collects events from source and routes to sink", func() {
			go nozzle.Start()

			time.Sleep(time.Second)

			Eventually(func() []*events.Envelope {
				return eventRouter.Events()
			}).Should(HaveLen(10))
		})

	})

	prepare := func(closeErr int) func() {
		return func() {
			consumer = testing.NewMockFirehoseConsumer(-1, int64(10), closeErr)
			eventRouter = testing.NewMockEventRouter()
			config := &FirehoseConfig{
				FirehoseSubscriptionID: "splunk-subcription-id",

				Logger: lager.NewLogger("test"),
			}
			nozzle = NewFirehoseNozzle(consumer, eventRouter, config)
		}
	}

	runAndAssert := func(closeErr int) func() {
		return func() {
			done := make(chan error, 1)
			go func() {
				err := nozzle.Start()
				done <- err
			}()

			time.Sleep(time.Second)
			nozzle.Close()

			err := <-done
			ce := err.(*websocket.CloseError)
			Expect(ce.Code).To(Equal(closeErr))
		}
	}

	Context("When there is websocket.CloseNormalClosure from event source", func() {
		BeforeEach(prepare(websocket.CloseNormalClosure))
		It("handles errors when collects events from source", runAndAssert(websocket.CloseNormalClosure))
	})

	Context("When there is websocket.ClosePolicyViolation from event source", func() {
		BeforeEach(prepare(websocket.ClosePolicyViolation))
		It("handles errors when collects events from source", runAndAssert(websocket.ClosePolicyViolation))
	})

	Context("When there is websocket.CloseGoingAway from event source", func() {
		BeforeEach(prepare(websocket.CloseGoingAway))
		It("handles errors when collects events from source", runAndAssert(websocket.CloseGoingAway))
	})
})
