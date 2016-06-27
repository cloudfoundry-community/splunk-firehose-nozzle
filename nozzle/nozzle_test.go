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
		nozzle Nozzle

		eventChannel        chan *events.Envelope
		errorChannel        chan error
		capturedSplunkEvent *SplunkEvent
	)

	BeforeEach(func() {
		eventChannel = make(chan *events.Envelope)
		errorChannel = make(chan error, 1)

		capturedSplunkEvent = nil
		nozzle = NewSplunkForwarder(&MockSplunkClient{
			PostFn: func(event *SplunkEvent) error {
				capturedSplunkEvent = event
				return nil
			},
		}, eventChannel, errorChannel)
	})

	It("returns error on error channel", func() {
		go func() {
			errorChannel <- errors.New("Fail")
		}()
		err := nozzle.Run()

		Expect(err).To(Equal(errors.New("Fail")))
	})

	Context("Envelope ValueMetric", func() {
		var origin, deployment, job, index, ip string
		var timestampNano int64
		var envelopeValueMetric *events.ValueMetric
		var eventType = events.Envelope_ValueMetric

		var name, unit string
		var value float64
		var envelope *events.Envelope

		BeforeEach(func() {
			name = "ms_since_last_registry_update"
			value = 1581.0
			unit = "ms"
			envelopeValueMetric = &events.ValueMetric{
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
				ValueMetric: envelopeValueMetric,
			}
		})

		It("posts envelope", func() {
			go func() { nozzle.Run() }()

			eventChannel <- envelope

			Expect(capturedSplunkEvent).NotTo(BeNil())
		})

		It("correctly translates splunk metadata", func() {
			metric := BuildValueMetric(envelope)

			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
		})

		It("correctly translates to splunk", func() {
			metric := BuildValueMetric(envelope)
			event := metric.Event.(SplunkValueMetric)

			Expect(event.Name).To(Equal(name))
			Expect(event.Value).To(Equal(value))
			Expect(event.Unit).To(Equal(unit))
		})
	})

	Context("Envelope Error", func() {
		var origin, deployment, job, index, ip string
		var timestampNano int64
		var envelopeError *events.Error
		var eventType = events.Envelope_ValueMetric

		var source, message string
		var code int32
		var envelope *events.Envelope

		BeforeEach(func() {
			source = "some_source"
			message = "something failed"
			code = 42
			envelopeError = &events.Error{
				Source:  &source,
				Code:    &code,
				Message: &message,
			}

			origin = "Unknown"
			eventType = events.Envelope_Error
			timestampNano = 1467040874046121775
			deployment = "cf-warden"
			job = "router_z1"
			index = "0"
			ip = "10.244.0.22"
			envelope = &events.Envelope{
				Origin:     &origin,
				EventType:  &eventType,
				Timestamp:  &timestampNano,
				Deployment: &deployment,
				Job:        &job,
				Index:      &index,
				Ip:         &ip,
				Error:      envelopeError,
			}
		})

		It("posts envelope", func() {
			go func() { nozzle.Run() }()

			eventChannel <- envelope

			Expect(capturedSplunkEvent).NotTo(BeNil())
		})

		It("correctly translates splunk metadata", func() {
			metric := BuildErrorMetric(envelope)

			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
		})

		It("correctly translates to splunk", func() {
			metric := BuildErrorMetric(envelope)
			event := metric.Event.(SplunkErrorMetric)

			Expect(event.Source).To(Equal(source))
			Expect(event.Code).To(Equal(code))
			Expect(event.Message).To(Equal(message))
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
