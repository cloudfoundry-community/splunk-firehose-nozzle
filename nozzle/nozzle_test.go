package nozzle_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/nozzle"
)

var _ = Describe("Nozzle", func() {
	var (
		eventChannel chan *events.Envelope
		errorChannel chan error
	)

	BeforeEach(func() {
		eventChannel = make(chan *events.Envelope)
		errorChannel = make(chan error, 1)
	})

	It("returns error on error channel", func() {
		nozzle := NewSplunkForwarder(&MockSplunkClient{}, eventChannel, errorChannel)
		go func() {
			errorChannel <- errors.New("Fail")
		}()
		err := nozzle.Run()

		Expect(err).To(Equal(errors.New("Fail")))
	})

	Context("ValueMetric", func() {
		var name, unit, origin, deployment, job, index, ip string
		var value float64
		var timestampNano int64

		var valueMetric *events.ValueMetric
		var eventType = events.Envelope_ValueMetric
		var envelope *events.Envelope

		BeforeEach(func() {
			name = "ms_since_last_registry_update"
			value = 1581.0
			unit = "ms"
			valueMetric = &events.ValueMetric{
				Name:  &name,
				Value: &value,
				Unit:  &unit,
			}

			origin = "MetronAgent"
			eventType = events.Envelope_ValueMetric
			timestampNano = 1467040874046121775
			deployment = "cf-warden"
			job = "router_z1"
			index = "0"
			ip = "10.244.0.22"
			envelope = &events.Envelope{
				Origin:      &origin,
				EventType:   &eventType,
				Timestamp:   &timestampNano,
				Deployment:  &deployment,
				Job:         &job,
				Index:       &index,
				Ip:          &ip,
				ValueMetric: valueMetric,
			}
		})

		It("posts envelope", func() {
			var capturedSplunkEvent *SplunkEvent
			nozzle := NewSplunkForwarder(&MockSplunkClient{
				PostFn: func(event *SplunkEvent) error {
					capturedSplunkEvent = event
					return nil
				},
			}, eventChannel, errorChannel)

			go func() {
				nozzle.Run()
			}()

			eventChannel <- envelope

			Expect(capturedSplunkEvent).NotTo(BeNil())
		})

		It("correctly translates splunk metadata", func() {
			metric := EventValueMetric(envelope)

			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
		})

		It("correctly translates splunk value metric", func() {
			metric := EventValueMetric(envelope)
			event := metric.Event.(SplunkValueMetric)

			Expect(event.Name).To(Equal(name))
			Expect(event.Value).To(Equal(value))
			Expect(event.Unit).To(Equal(unit))
		})
	})
})

type MockSplunkClient struct {
	PostFn func(event *SplunkEvent) error
}

func (m *MockSplunkClient) Post(event *SplunkEvent) error {
	if m.PostFn != nil {
		return m.PostFn(event)
	}
	return nil
}
