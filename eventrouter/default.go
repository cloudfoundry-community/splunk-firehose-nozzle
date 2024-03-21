package eventrouter

import (
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	fevents "github.com/cloudfoundry-community/splunk-firehose-nozzle/events"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsink"
	"github.com/cloudfoundry/sonde-go/events"
)

type Config = fevents.Config

type router struct {
	appCache       cache.Cache
	sink           eventsink.Sink
	selectedEvents map[string]bool
	config         *Config
}

func New(appCache cache.Cache, sink eventsink.Sink, config *Config) (Router, error) {
	selectedEvents, err := fevents.ParseSelectedEvents(config.SelectedEvents)

	if err != nil {
		return nil, err
	}

	return &router{
		appCache:       appCache,
		sink:           sink,
		selectedEvents: selectedEvents,
		config:         config,
	}, nil
}

func (r *router) Route(msg *events.Envelope) error {
	eventType := msg.GetEventType()

	if _, ok := r.selectedEvents[eventType.String()]; !ok {
		// Ignore this event since we are not interested
		return nil
	}
	_ = r.sink.Write(msg)

	return nil
}
