package nozzle

import (
	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventrouter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsource"
)

// Config struct with field type Logger
type Config struct {
	Logger lager.Logger
}

// Nozzle struct with fields to read and route data.
type Nozzle struct {
	eventSource eventsource.Source
	eventRouter eventrouter.Router
	config      *Config

	closing chan struct{}
	closed  chan struct{}
}

// New returns Nozzle which reads events from eventsource.Source and routes events
// to targets by using eventrouter.Router
func New(eventSource eventsource.Source, eventRouter eventrouter.Router, config *Config) *Nozzle {
	return &Nozzle{
		eventRouter: eventRouter,
		eventSource: eventSource,
		config:      config,
		closing:     make(chan struct{}, 1),
		closed:      make(chan struct{}, 1),
	}
}

// Start initiates the nozzle and start reading events
func (f *Nozzle) Start() error {
	err := f.eventSource.Open()
	if err != nil {
		return err
	}

	defer close(f.closed)

	var lastErr error
	events := f.eventSource.Read()
	for {
		select {
		case event, ok := <-events:
			if !ok {
				f.config.Logger.Info("Give up after retries. Firehose consumer is going to exit")
				return lastErr
			}

			if err := f.eventRouter.Route(event); err != nil {
				f.config.Logger.Error("Failed to route event", err)
			}

		case <-f.closing:
			return lastErr
		}
	}
}

// Close closes the eventSource connection and stops reading events
func (f *Nozzle) Close() error {
	err := f.eventSource.Close()
	if err != nil {
		return err
	}

	close(f.closing)
	<-f.closed
	return nil
}
