package nozzle

import (
	"fmt"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/pivotal-golang/lager"
)

type Nozzle interface {
	Run(flushWindow time.Duration) error
}

type ForwardingNozzle struct {
	client             Client
	eventSerializer    EventSerializer
	includedEventTypes map[events.Envelope_EventType]bool
	eventsChannel      <-chan *events.Envelope
	errorsChannel      <-chan error
	batch              []interface{}
	logger             lager.Logger
}

type Client interface {
	PostBatch([]interface{}) error
}

type EventSerializer interface {
	BuildHttpStartStopEvent(event *events.Envelope) interface{}
	BuildLogMessageEvent(event *events.Envelope) interface{}
	BuildValueMetricEvent(event *events.Envelope) interface{}
	BuildCounterEvent(event *events.Envelope) interface{}
	BuildErrorEvent(event *events.Envelope) interface{}
	BuildContainerEvent(event *events.Envelope) interface{}
}

func NewForwarder(clientlient Client, eventSerializer EventSerializer, selectedEventTypes []events.Envelope_EventType, eventsChannel <-chan *events.Envelope, errors <-chan error, logger lager.Logger) Nozzle {
	nozzle := &ForwardingNozzle{
		client:          clientlient,
		eventSerializer: eventSerializer,
		eventsChannel:   eventsChannel,
		errorsChannel:   errors,
		batch:           make([]interface{}, 0),
		logger:          logger,
	}

	nozzle.includedEventTypes = map[events.Envelope_EventType]bool{
		events.Envelope_HttpStart:       false,
		events.Envelope_HttpStop:        false,
		events.Envelope_HttpStartStop:   false,
		events.Envelope_LogMessage:      false,
		events.Envelope_ValueMetric:     false,
		events.Envelope_CounterEvent:    false,
		events.Envelope_Error:           false,
		events.Envelope_ContainerMetric: false,
	}
	for _, selectedEventType := range selectedEventTypes {
		nozzle.includedEventTypes[selectedEventType] = true
	}

	return nozzle
}

func (s *ForwardingNozzle) Run(flushWindow time.Duration) error {
	ticker := time.Tick(flushWindow)
	for {
		select {
		case err := <-s.errorsChannel:
			return err
		case event := <-s.eventsChannel:
			s.handleEvent(event)
		case <-ticker:
			if len(s.batch) > 0 {
				s.logger.Info(fmt.Sprintf("Posting %d events", len(s.batch)))
				s.client.PostBatch(s.batch)
				s.batch = make([]interface{}, 0)
			} else {
				s.logger.Info(fmt.Sprintf("No events to post"))
			}
		}
	}
}

func (s *ForwardingNozzle) handleEvent(envelope *events.Envelope) {
	var event interface{} = nil

	eventType := envelope.GetEventType()
	if !s.includedEventTypes[eventType] {
		return
	}

	switch eventType {
	case events.Envelope_HttpStart:
	case events.Envelope_HttpStop:
	case events.Envelope_HttpStartStop:
		event = s.eventSerializer.BuildHttpStartStopEvent(envelope)
	case events.Envelope_LogMessage:
		event = s.eventSerializer.BuildLogMessageEvent(envelope)
	case events.Envelope_ValueMetric:
		event = s.eventSerializer.BuildValueMetricEvent(envelope)
	case events.Envelope_CounterEvent:
		event = s.eventSerializer.BuildCounterEvent(envelope)
	case events.Envelope_Error:
		event = s.eventSerializer.BuildErrorEvent(envelope)
	case events.Envelope_ContainerMetric:
		event = s.eventSerializer.BuildContainerEvent(envelope)
	}

	if event != nil {
		s.batch = append(s.batch, event)
	}
}
