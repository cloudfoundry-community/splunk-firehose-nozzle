package nozzle

import (
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)

type Nozzle interface {
	Run() error
}

type SplunkNozzle struct {
	events <-chan *events.Envelope
	errors <-chan error
}

func NewSplunkForwarder(events <-chan *events.Envelope, errors <-chan error) Nozzle {
	return &SplunkNozzle{
		events: events,
		errors: errors,
	}
}

func (s *SplunkNozzle) Run() error {
	for {
		select {
		case err := <-s.errors:
			return err
		case event := <-s.events:
			s.handleEvent(event)
		}
	}
}

func (s *SplunkNozzle) handleEvent(event *events.Envelope) {
	//todo: exploratory work, delete & tdd actual solution
	eventType := event.EventType
	if *eventType == events.Envelope_LogMessage {
		logMessage := event.LogMessage
		message := string(logMessage.Message)
		fmt.Printf("%s", message)
	}
}
