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

		origin        string
		deployment    string
		job           string
		index         string
		ip            string
		timestampNano int64
		envelope      *events.Envelope
		eventType     events.Envelope_EventType
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

		timestampNano = 1467040874046121775
		deployment = "cf-warden"
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
		}
	})

	It("returns error on error channel", func() {
		go func() {
			errorChannel <- errors.New("Fail")
		}()
		err := nozzle.Run()

		Expect(err).To(Equal(errors.New("Fail")))
	})

	Context("Envelope LogMessage", func() {
		var message []byte
		var messageType events.LogMessage_MessageType
		var timestamp int64
		var appId, sourceType, sourceInstance string
		var envelopeLogMessage *events.LogMessage

		BeforeEach(func() {
			message = []byte("App debug log message")
			messageType = events.LogMessage_OUT
			timestamp = 1467128185055072010
			appId = "8463ec45-543c-4492-9ec6-f52707f7dd2b"
			sourceType = "App"
			sourceInstance = "0"
			envelopeLogMessage = &events.LogMessage{
				Message:        message,
				MessageType:    &messageType,
				Timestamp:      &timestamp,
				AppId:          &appId,
				SourceType:     &sourceType,
				SourceInstance: &sourceInstance,
			}

			job = "runner_z1"
			origin = "dea_logging_agent"
			eventType = events.Envelope_LogMessage
			envelope.LogMessage = envelopeLogMessage
		})

		It("posts envelope", func() {
			go func() { nozzle.Run() }()

			eventChannel <- envelope

			Eventually(func() *SplunkEvent {
				return capturedSplunkEvent
			}).ShouldNot(BeNil())
			Expect(
				capturedSplunkEvent.Event.(SplunkLogMessageMetric).EventType,
			).To(Equal("LogMessage"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildLogMessageMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkLogMessageMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("LogMessage"))
			})

			It("log message specific data", func() {
				event := metric.Event.(SplunkLogMessageMetric)

				Expect(event.Message).To(Equal(string(message)))
				Expect(event.MessageType).To(Equal("OUT"))
				Expect(event.Timestamp).To(Equal("1467128185.055"))
				Expect(event.AppId).To(Equal(appId))
				Expect(event.SourceType).To(Equal(sourceType))
				Expect(event.SourceInstance).To(Equal(sourceInstance))
			})
		})
	})

	Context("Envelope ValueMetric", func() {
		var name, unit string
		var value float64
		var envelopeValueMetric *events.ValueMetric

		BeforeEach(func() {
			name = "ms_since_last_registry_update"
			value = 1581.0
			unit = "ms"
			envelopeValueMetric = &events.ValueMetric{
				Name:  &name,
				Value: &value,
				Unit:  &unit,
			}

			job = "router_z1"
			origin = "MetronAgent"
			eventType = events.Envelope_ValueMetric
			envelope.ValueMetric = envelopeValueMetric
		})

		It("posts envelope", func() {
			go func() { nozzle.Run() }()

			eventChannel <- envelope

			Eventually(func() *SplunkEvent {
				return capturedSplunkEvent
			}).ShouldNot(BeNil())
			Expect(
				capturedSplunkEvent.Event.(SplunkValueMetric).EventType,
			).To(Equal("ValueMetric"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildValueMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkValueMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("ValueMetric"))
			})

			It("metric specific data", func() {
				event := metric.Event.(SplunkValueMetric)

				Expect(event.Name).To(Equal(name))
				Expect(event.Value).To(Equal(value))
				Expect(event.Unit).To(Equal(unit))
			})
		})
	})

	Context("Envelope CounterEvent", func() {
		var name string
		var delta, total uint64
		var counterEvent *events.CounterEvent

		BeforeEach(func() {
			name = "registry_message.uaa"
			delta = 1
			total = 8196
			counterEvent = &events.CounterEvent{
				Name:  &name,
				Delta: &delta,
				Total: &total,
			}

			job = "router_z1"
			origin = "gorouter"
			eventType = events.Envelope_CounterEvent
			envelope.CounterEvent = counterEvent
		})

		It("posts envelope", func() {
			go func() { nozzle.Run() }()

			eventChannel <- envelope

			Eventually(func() *SplunkEvent {
				return capturedSplunkEvent
			}).ShouldNot(BeNil())
			Expect(
				capturedSplunkEvent.Event.(SplunkCounterEventMetric).EventType,
			).To(Equal("CounterEvent"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildCounterEventMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkCounterEventMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("CounterEvent"))
			})

			It("metric specific data", func() {
				event := metric.Event.(SplunkCounterEventMetric)

				Expect(event.Name).To(Equal(name))
				Expect(event.Delta).To(Equal(delta))
				Expect(event.Total).To(Equal(total))
			})
		})
	})

	Context("Envelope Error", func() {
		var source, message string
		var code int32
		var envelopeError *events.Error

		BeforeEach(func() {
			source = "some_source"
			message = "something failed"
			code = 42
			envelopeError = &events.Error{
				Source:  &source,
				Code:    &code,
				Message: &message,
			}

			job = "router_z1"
			origin = "Unknown"
			eventType = events.Envelope_Error
			envelope.Error = envelopeError
		})

		It("posts envelope", func() {
			go func() { nozzle.Run() }()

			eventChannel <- envelope

			Eventually(func() *SplunkEvent {
				return capturedSplunkEvent
			}).ShouldNot(BeNil())
			Expect(
				capturedSplunkEvent.Event.(SplunkErrorMetric).EventType,
			).To(Equal("Error"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildErrorMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkErrorMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("Error"))
			})

			It("metric specific data", func() {
				event := metric.Event.(SplunkErrorMetric)

				Expect(event.Source).To(Equal(source))
				Expect(event.Code).To(Equal(code))
				Expect(event.Message).To(Equal(message))
			})
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
