package eventsource

import (
	"github.com/cloudfoundry/sonde-go/events"
)

//go:generate counterfeiter . Source

type Source interface {
	Open() error
	Close() error
	Read() (<-chan *events.Envelope, <-chan error)
}
