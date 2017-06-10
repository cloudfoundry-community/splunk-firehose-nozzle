package firehoseclient_test

import (
	"time"

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
			consumer = testing.NewMockFirehoseConsumer(10, int64(10), -1)
			eventRouter = testing.NewMockEventRouter()
			config := &FirehoseConfig{
				TrafficControllerURL:   "https://api.bosh-lite.com",
				InsecureSSLSkipVerify:  true,
				IdleTimeoutSeconds:     time.Second,
				FirehoseSubscriptionID: "splunk-subcription-id",
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

	closeErrs := []int{websocket.CloseNormalClosure, websocket.ClosePolicyViolation, websocket.CloseGoingAway}
	for _, closeErr := range closeErrs {
		Context("When there are errors from event source", func() {
			BeforeEach(func() {
				consumer = testing.NewMockFirehoseConsumer(10, int64(10), closeErr)
				eventRouter = testing.NewMockEventRouter()
				config := &FirehoseConfig{
					TrafficControllerURL:   "https://api.bosh-lite.com",
					InsecureSSLSkipVerify:  true,
					IdleTimeoutSeconds:     time.Second,
					FirehoseSubscriptionID: "splunk-subcription-id",
				}
				nozzle = NewFirehoseNozzle(consumer, eventRouter, config)
			})

			It("handles errors when collects events from source", func() {
				done := make(chan error, 1)
				go func() {
					err := nozzle.Start()
					done <- err
				}()

				time.Sleep(time.Second)

				err := <-done
				ce := err.(*websocket.CloseError)
				Expect(ce.Code).To(Equal(closeErr))
			})
		})

	}
})
