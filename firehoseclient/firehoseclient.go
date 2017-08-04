package firehoseclient

import (
	"code.cloudfoundry.org/lager"

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
	consumer    FirehoseConsumer
	eventRouter EventRouter
	config      *FirehoseConfig

	closing chan struct{}
	closed  chan struct{}
}

type FirehoseConfig struct {
	Logger lager.Logger

	FirehoseSubscriptionID string
}

func NewFirehoseNozzle(consumer FirehoseConsumer, eventRouter EventRouter, firehoseconfig *FirehoseConfig) *FirehoseNozzle {
	return &FirehoseNozzle{
		eventRouter: eventRouter,
		consumer:    consumer,
		config:      firehoseconfig,
		closing:     make(chan struct{}, 1),
		closed:      make(chan struct{}, 1),
	}
}

func (f *FirehoseNozzle) Start() error {
	defer close(f.closed)

	var lastErr error
	messages, errs := f.consumer.Firehose(f.config.FirehoseSubscriptionID, "")
	for {
		select {
		case envelope, ok := <-messages:
			if !ok {
				f.config.Logger.Info("Give up after retries. Firehose consumer is going to exit")
				return lastErr
			}

			f.eventRouter.RouteEvent(envelope)

		case lastErr = <-errs:
			f.handleError(lastErr)

		case <-f.closing:
			return lastErr
		}
	}
}

func (f *FirehoseNozzle) Close() error {
	close(f.closing)
	<-f.closed
	return nil
}

func (f *FirehoseNozzle) handleError(err error) {
	closeErr, ok := err.(*websocket.CloseError)
	if !ok {
		f.config.Logger.Error("Error while reading from the firehose", err)
		return
	}

	msg := ""
	switch closeErr.Code {
	case websocket.CloseNormalClosure:
		msg = "Connection was disconnected by Firehose server. This usually means Nozzle can't keep up " +
			"with server. Please try to scaling out Nozzzle with mulitple instances by using the " +
			"same subscription ID."

	case websocket.ClosePolicyViolation:
		msg = "Nozzle lost the keep-alive heartbeat with Firehose server. Connection was disconnected " +
			"by Firehose server. This usually means either Nozzle was busy with processing events or there " +
			"was some temperary network issue causing the heartbeat to get lost."

	default:
		msg = "Encountered close error while reading from Firehose"
	}

	f.config.Logger.Error(msg, err)
}
