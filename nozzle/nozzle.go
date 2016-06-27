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
	var splunkEvent *SplunkEvent = nil

	switch *event.EventType {
	case events.Envelope_HttpStart:
	case events.Envelope_HttpStop:
	case events.Envelope_HttpStartStop:
	case events.Envelope_LogMessage:
	case events.Envelope_ValueMetric:
		splunkEvent = BuildValueMetric(event)
	case events.Envelope_CounterEvent:
	case events.Envelope_Error:
		splunkEvent = BuildErrorMetric(event)
	case events.Envelope_ContainerMetric:
	}

	if splunkEvent != nil {
		s.splunkClient.Post(splunkEvent)
	}
}

type SplunkErrorMetric struct {
	Source  string `json:"source"`
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func BuildErrorMetric(nozzleEvent *events.Envelope) *SplunkEvent {
	errorMetric := nozzleEvent.Error
	splunkErrorMetric := SplunkErrorMetric{
		Source:  errorMetric.GetSource(),
		Code:    errorMetric.GetCode(),
		Message: errorMetric.GetMessage(),
	}

	splunkEvent := &SplunkEvent{
		Time:   nanoSecondsToSeconds(nozzleEvent.GetTimestamp()),
		Host:   nozzleEvent.GetIp(),
		Source: nozzleEvent.GetJob(),
		Event:  splunkErrorMetric,
	}
	return splunkEvent
}

type SplunkValueMetric struct {
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func BuildValueMetric(nozzleEvent *events.Envelope) *SplunkEvent {
	valueMetric := nozzleEvent.ValueMetric
	splunkValueMetric := SplunkValueMetric{
		Name:  valueMetric.GetName(),
		Value: valueMetric.GetValue(),
		Unit:  valueMetric.GetUnit(),
	}

	splunkEvent := &SplunkEvent{
		Time:   nanoSecondsToSeconds(nozzleEvent.GetTimestamp()),
		Host:   nozzleEvent.GetIp(),
		Source: nozzleEvent.GetJob(), //todo: consider app vs cf once understand full metric set
		Event:  splunkValueMetric,
	}
	return splunkEvent
}

func nanoSecondsToSeconds(nanoseconds int64) string {
	seconds := float64(nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.3f", seconds)
}
