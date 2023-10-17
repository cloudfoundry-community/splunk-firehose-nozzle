package eventsink

import "github.com/cloudfoundry/sonde-go/events"

type Sink interface {
	Open() error
	Close() error
	Write(fields *events.Envelope) error
}
