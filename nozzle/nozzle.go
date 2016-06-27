package nozzle

import (
	"fmt"
	"math"

	"github.com/cloudfoundry/sonde-go/events"
)

type Nozzle interface {
	Run() error
}

type SplunkNozzle struct {
	splunkClient SplunkClient
	events       <-chan *events.Envelope
	errors       <-chan error
}

func NewSplunkForwarder(splunkClient SplunkClient, events <-chan *events.Envelope, errors <-chan error) Nozzle {
	return &SplunkNozzle{
		splunkClient: splunkClient,
		events:       events,
		errors:       errors,
	}
}

func (s *SplunkNozzle) Run() error {
	go func() {
		for event := range s.events {
			s.handleEvent(event)
		}
	}()

	return <-s.errors
}

func (s *SplunkNozzle) handleEvent(event *events.Envelope) {
	eventType := *event.EventType

	switch eventType {
	case events.Envelope_HttpStart:
	case events.Envelope_HttpStop:
	case events.Envelope_HttpStartStop:
	case events.Envelope_LogMessage:
	case events.Envelope_ValueMetric:
		metric := EventValueMetric(event)
		s.splunkClient.Post(metric)
	case events.Envelope_CounterEvent:
	case events.Envelope_Error:
	case events.Envelope_ContainerMetric:
	}
}

type SplunkValueMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func EventValueMetric(nozzleEvent *events.Envelope) *SplunkEvent {
	valueMetric := nozzleEvent.ValueMetric
	splunkValueMetric := SplunkValueMetric{
		Name:  *valueMetric.Name,
		Value: *valueMetric.Value,
		Unit:  *valueMetric.Unit,
	}

	splunkEvent := &SplunkEvent{
		Time:   nanoSecondsToSeconds(nozzleEvent.Timestamp),
		Host:   *nozzleEvent.Ip,
		Source: *nozzleEvent.Job, //todo: consider app vs cf once understand full metric set
		Event:  splunkValueMetric,
	}
	return splunkEvent
}

func nanoSecondsToSeconds(nanoseconds *int64) string {
	seconds := float64(*nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.3f", seconds)
}
