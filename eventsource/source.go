package eventsource

import (
	"github.com/cloudfoundry/sonde-go/events"
)

//go:generate counterfeiter . Source

// Source provides a mechanism to initiate Open, Close and Read methods.
type Source interface {
	Open() error
	Close() error
	Read() <-chan *events.Envelope
}
