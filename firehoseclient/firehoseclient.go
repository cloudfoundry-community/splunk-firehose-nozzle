package firehoseclient

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
)

type FirehoseConsumer interface {
	Firehose(subscriptionId string, authToken string) (<-chan *events.Envelope, <-chan error)
	Close() error
}

type EventRouter interface {
	RouteEvent(msg *events.Envelope)
}

type FirehoseNozzle struct {
	consumer     FirehoseConsumer
	eventRouting EventRouter
	config       *FirehoseConfig
}

type FirehoseConfig struct {
	FirehoseSubscriptionID string
}

func NewFirehoseNozzle(consumer FirehoseConsumer, eventRouting EventRouter, firehoseconfig *FirehoseConfig) *FirehoseNozzle {
	return &FirehoseNozzle{
		eventRouting: eventRouting,
		consumer:     consumer,
		config:       firehoseconfig,
	}
}

func (f *FirehoseNozzle) Start() error {
	return f.routeEvent()
}

func (f *FirehoseNozzle) routeEvent() error {
	messages, errs := f.consumer.Firehose(f.config.FirehoseSubscriptionID, "")
	for {
		select {
		case envelope := <-messages:
			f.eventRouting.RouteEvent(envelope)
		case err := <-errs:
			f.handleError(err)
			return err
		}
	}
}

func (f *FirehoseNozzle) handleError(err error) {
	switch {
	case websocket.IsCloseError(err, websocket.CloseNormalClosure):
		logging.LogError("Normal Websocket Closure", err)
	case websocket.IsCloseError(err, websocket.ClosePolicyViolation):
		logging.LogError("Error while reading from the firehose", err)
		logging.LogError("Disconnected because nozzle couldn't keep up. Please try scaling up the nozzle.", nil)

	default:
		logging.LogError("Error while reading from the firehose", err)
	}

	logging.LogError("Closing connection with traffic controller due to", err)
	f.consumer.Close()
}
