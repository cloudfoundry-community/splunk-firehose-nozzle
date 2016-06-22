package nozzle

import (
	"fmt"

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
		fmt.Printf("Posting %s", message)

		err := s.splunkClient.Post(&SplunkEvent{
			Event: message,
		})
		if err != nil {
			println(fmt.Sprintf("Error posting to splunk: %s", err.Error()))
		}
	}
}
