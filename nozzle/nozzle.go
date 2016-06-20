package nozzle

import (
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)

type Forwarder interface {
	Forward() error
}

type SplunkNozzle struct {
	events <-chan *events.Envelope
	errors <-chan error
}

func NewSplunkForwarder(events <-chan *events.Envelope, errors <-chan error) Forwarder {
	return &SplunkNozzle{
		events: events,
		errors: errors,
	}
}

func (s *SplunkNozzle) Forward() error {
	for {
		select {
		case err := <-s.errors:
			return err
		case event := <-s.events:
			fmt.Printf("%+v\n", *event)
		}
	}
}
