package splunknozzle

import (
	"github.com/cloudfoundry/sonde-go/events"
)

type FirehoseConsumer interface {
	Firehose(subscriptionId string, authToken string) (<-chan *events.Envelope, <-chan error)
	Close() error
}

type EventRouter interface {
	RouteEvent(msg *events.Envelope)
}
