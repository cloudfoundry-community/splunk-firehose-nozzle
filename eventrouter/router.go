package eventrouter

import (
	"github.com/cloudfoundry/sonde-go/events"
)

type Router interface {
	Route(msg *events.Envelope) error
}
