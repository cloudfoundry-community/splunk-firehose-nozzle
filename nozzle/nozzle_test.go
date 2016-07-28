package nozzle_test

import (
	"bytes"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/nozzle"
)

var _ = Describe("Nozzle", func() {
	var (
		nozzle Nozzle
		logger lager.Logger

		allEventTypes       []events.Envelope_EventType
		eventChannel        chan *events.Envelope
		errorChannel        chan error
		capturedEvents      []interface{}
		mockEventSerializer *MockEventSerializer

		origin        string
		deployment    string
		job           string
		index         string
		ip            string
		timestampNano int64
		envelope      *events.Envelope
		eventType     events.Envelope_EventType
	)

	var newEvent = func(e events.Envelope_EventType) *events.Envelope {
		newEventType := e
		return &events.Envelope{
			Origin:     &origin,
			EventType:  &newEventType,
			Timestamp:  &timestampNano,
			Deployment: &deployment,
			Job:        &job,
			Index:      &index,
			Ip:         &ip,
		}
	}

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		eventChannel = make(chan *events.Envelope)
		errorChannel = make(chan error, 1)
		allEventTypes = []events.Envelope_EventType{
			events.Envelope_HttpStartStop,
			events.Envelope_LogMessage,
			events.Envelope_ValueMetric,
			events.Envelope_CounterEvent,
			events.Envelope_Error,
			events.Envelope_ContainerMetric,
		}

		capturedEvents = nil
		mockClient := &MockClient{
			PostBatchFn: func(events []interface{}) error {
				capturedEvents = events
				return nil
			},
		}
		mockEventSerializer = NewMockEventSerializer()

		nozzle = NewForwarder(
			mockClient, mockEventSerializer, allEventTypes, eventChannel, errorChannel, logger,
		)

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

	It("logs when error on channel", func() {
		buff := &bytes.Buffer{}
		logger.RegisterSink(lager.NewWriterSink(buff, lager.ERROR))

		go func() {
			errorChannel <- errors.New("Fail")
		}()
		var runErr error
		go func() {
			runErr = nozzle.Run(time.Millisecond * 100)
		}()

		Consistently(func() error {
			return runErr
		}).Should(BeNil())
		Expect(buff.Len()).To(BeNumerically(">", 0))
		Expect(buff.String()).To(ContainSubstring("firehose"))
	})

	It("returns error when error channel closed", func() {
		var runErr error
		go func() {
			runErr = nozzle.Run(time.Millisecond * 100)
		}()
		close(errorChannel)

		Eventually(func() error {
			return runErr
		}).ShouldNot(BeNil())
		Expect(runErr.Error()).To(ContainSubstring("closed"))
	})

	It("returns error when event channel closed", func() {
		var runErr error
		go func() {
			runErr = nozzle.Run(time.Millisecond * 100)
		}()
		close(eventChannel)

		Eventually(func() error {
			return runErr
		}).ShouldNot(BeNil())
		Expect(runErr.Error()).To(ContainSubstring("closed"))
	})

	It("returns error when client fails", func() {
		mockClient := &MockClient{
			PostBatchFn: func(events []interface{}) error {
				return errors.New("failed to post")
			},
		}

		nozzle = NewForwarder(
			mockClient, mockEventSerializer, allEventTypes, eventChannel, errorChannel, logger,
		)

		var runErr error
		go func() {
			runErr = nozzle.Run(time.Millisecond * 100)
		}()

		envelope = newEvent(events.Envelope_HttpStartStop)
		envelope.HttpStartStop = &events.HttpStartStop{}
		eventChannel <- envelope

		Eventually(func() error {
			return runErr
		}).ShouldNot(BeNil())
		Expect(runErr.Error()).To(Equal("failed to post"))
	})

	Context("Filters events", func() {
		var sendAllEventTypes = func() {
			envelope = newEvent(events.Envelope_HttpStartStop)
			envelope.HttpStartStop = &events.HttpStartStop{}
			eventChannel <- envelope

			envelope = newEvent(events.Envelope_LogMessage)
			envelope.LogMessage = &events.LogMessage{}
			eventChannel <- envelope

			envelope = newEvent(events.Envelope_ValueMetric)
			envelope.ValueMetric = &events.ValueMetric{}
			eventChannel <- envelope

			envelope = newEvent(events.Envelope_CounterEvent)
			envelope.CounterEvent = &events.CounterEvent{}
			eventChannel <- envelope

			envelope = newEvent(events.Envelope_Error)
			envelope.Error = &events.Error{}
			eventChannel <- envelope

			envelope = newEvent(events.Envelope_ContainerMetric)
			envelope.ContainerMetric = &events.ContainerMetric{}
			eventChannel <- envelope
		}

		It("all events", func() {
			capturedEvents = make([]interface{}, 0)

			mockClient := &MockClient{
				PostBatchFn: func(events []interface{}) error {
					capturedEvents = append(capturedEvents, events...)
					return nil
				},
			}
			nozzle = NewForwarder(
				mockClient, mockEventSerializer, allEventTypes, eventChannel, errorChannel, logger,
			)

			go func() { nozzle.Run(time.Millisecond * 100) }()

			sendAllEventTypes()

			Eventually(func() []interface{} {
				return capturedEvents
			}).Should(HaveLen(6))

			Expect(mockEventSerializer.StartStopEvents).To(HaveLen(1))
			Expect(mockEventSerializer.LogMessageEvents).To(HaveLen(1))
			Expect(mockEventSerializer.ValueMetricEvents).To(HaveLen(1))
			Expect(mockEventSerializer.CounterEvents).To(HaveLen(1))
			Expect(mockEventSerializer.ErrorEvents).To(HaveLen(1))
			Expect(mockEventSerializer.ContainerEvents).To(HaveLen(1))
		})

		It("filters events (test single type)", func() {
			selectedEventTypes := []events.Envelope_EventType{events.Envelope_ValueMetric}

			capturedEvents = make([]interface{}, 0)

			mockClient := &MockClient{
				PostBatchFn: func(events []interface{}) error {
					capturedEvents = append(capturedEvents, events...)
					return nil
				},
			}
			nozzle = NewForwarder(
				mockClient, mockEventSerializer, selectedEventTypes, eventChannel, errorChannel, logger,
			)

			go func() { nozzle.Run(time.Millisecond * 100) }()

			sendAllEventTypes()

			Eventually(func() []interface{} {
				return capturedEvents
			}).Should(HaveLen(1))

			Expect(mockEventSerializer.StartStopEvents).To(HaveLen(0))
			Expect(mockEventSerializer.LogMessageEvents).To(HaveLen(0))
			Expect(mockEventSerializer.ValueMetricEvents).To(HaveLen(1))
			Expect(mockEventSerializer.CounterEvents).To(HaveLen(0))
			Expect(mockEventSerializer.ErrorEvents).To(HaveLen(0))
			Expect(mockEventSerializer.ContainerEvents).To(HaveLen(0))
		})

		It("filters events (test multiple types)", func() {
			selectedEventTypes := []events.Envelope_EventType{
				events.Envelope_ValueMetric,
				events.Envelope_CounterEvent,
				events.Envelope_ContainerMetric,
			}

			capturedEvents = make([]interface{}, 0)

			mockClient := &MockClient{
				PostBatchFn: func(events []interface{}) error {
					capturedEvents = append(capturedEvents, events...)
					return nil
				},
			}
			nozzle = NewForwarder(
				mockClient, mockEventSerializer, selectedEventTypes, eventChannel, errorChannel, logger,
			)

			go func() { nozzle.Run(time.Millisecond * 100) }()

			sendAllEventTypes()

			Eventually(func() []interface{} {
				return capturedEvents
			}).Should(HaveLen(3))

			Expect(mockEventSerializer.StartStopEvents).To(HaveLen(0))
			Expect(mockEventSerializer.LogMessageEvents).To(HaveLen(0))
			Expect(mockEventSerializer.ValueMetricEvents).To(HaveLen(1))
			Expect(mockEventSerializer.CounterEvents).To(HaveLen(1))
			Expect(mockEventSerializer.ErrorEvents).To(HaveLen(0))
			Expect(mockEventSerializer.ContainerEvents).To(HaveLen(1))
		})
	})
})

type MockClient struct {
	PostSingleFn func(event interface{}) error
	PostBatchFn  func(event []interface{}) error
}

func (m *MockClient) PostSingle(event interface{}) error {
	if m.PostSingleFn != nil {
		return m.PostSingleFn(event)
	}
	return nil
}

func (m *MockClient) PostBatch(event []interface{}) error {
	if m.PostBatchFn != nil {
		return m.PostBatchFn(event)
	}
	return nil
}

type MockEventSerializer struct {
	StartStopEvents   []*events.Envelope
	LogMessageEvents  []*events.Envelope
	ValueMetricEvents []*events.Envelope
	CounterEvents     []*events.Envelope
	ErrorEvents       []*events.Envelope
	ContainerEvents   []*events.Envelope
}

func (m *MockEventSerializer) BuildHttpStartStopEvent(event *events.Envelope) interface{} {
	m.StartStopEvents = append(m.StartStopEvents, event)
	return event.GetEventType().String()
}

func (m *MockEventSerializer) BuildLogMessageEvent(event *events.Envelope) interface{} {
	m.LogMessageEvents = append(m.LogMessageEvents, event)
	return event.GetEventType().String()
}

func (m *MockEventSerializer) BuildValueMetricEvent(event *events.Envelope) interface{} {
	m.ValueMetricEvents = append(m.ValueMetricEvents, event)
	return event.GetEventType().String()
}

func (m *MockEventSerializer) BuildCounterEvent(event *events.Envelope) interface{} {
	m.CounterEvents = append(m.CounterEvents, event)
	return event.GetEventType().String()
}

func (m *MockEventSerializer) BuildErrorEvent(event *events.Envelope) interface{} {
	m.ErrorEvents = append(m.ErrorEvents, event)
	return event.GetEventType().String()
}

func (m *MockEventSerializer) BuildContainerEvent(event *events.Envelope) interface{} {
	m.ContainerEvents = append(m.ContainerEvents, event)
	return event.GetEventType().String()
}

func NewMockEventSerializer() *MockEventSerializer {
	return &MockEventSerializer{
		StartStopEvents:   make([]*events.Envelope, 0),
		LogMessageEvents:  make([]*events.Envelope, 0),
		ValueMetricEvents: make([]*events.Envelope, 0),
		CounterEvents:     make([]*events.Envelope, 0),
		ErrorEvents:       make([]*events.Envelope, 0),
		ContainerEvents:   make([]*events.Envelope, 0),
	}
}
