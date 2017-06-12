package testing

import (
	"fmt"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
)

type MockFirehoseConsumer struct {
	events      chan *events.Envelope
	errors      chan error
	eps         int
	totalEvents int64
	done        chan struct{}
}

func NewMockFirehoseConsumer(eps int, totalEvents int64, errCode int) *MockFirehoseConsumer {
	consumer := &MockFirehoseConsumer{
		events:      make(chan *events.Envelope, eps+1),
		errors:      make(chan error, 1),
		eps:         eps,
		totalEvents: totalEvents,
		done:        make(chan struct{}, 1),
	}

	// If generates error
	if errCode > 0 {
		err := &websocket.CloseError{
			Code: errCode,
		}

		consumer.errors <- err
	}

	go consumer.publishEvents()
	return consumer
}

func (consumer *MockFirehoseConsumer) Firehose(subscriptionId string, authToken string) (<-chan *events.Envelope, <-chan error) {
	return consumer.events, consumer.errors
}

func (consumer *MockFirehoseConsumer) Close() error {
	var done struct{}
	consumer.done <- done
	<-consumer.done
	return nil
}

func (consumer *MockFirehoseConsumer) publishEvents() {
	durationPerEvent := time.Duration(int64(time.Second) / int64(consumer.eps))
	tickerChan := time.NewTicker(durationPerEvent).C
	eventSent := int64(0)

	var (
		origin     = "DopplerServer"
		deployment = "cf"
		job        = "doppler"
		index      = "5a634d0b-bbc5-47c4-9450-a43f44a7fd30"
		ip         = "192.168.16.26"

		eventType    = events.Envelope_ValueMetric
		timestamp    = int64(0)
		tags         = map[string]string{}
		metricName   = "messageRouter.numberOfFirehoseSinks"
		metricValue  = float64(1)
		metricUnit   = "sinks"
		unrecognized = []byte{}
		metric       = events.ValueMetric{
			Name:             &metricName,
			Value:            &metricValue,
			Unit:             &metricUnit,
			XXX_unrecognized: unrecognized,
		}
	)

	event := &events.Envelope{
		Origin:      &origin,
		Deployment:  &deployment,
		Job:         &job,
		Index:       &index,
		Ip:          &ip,
		EventType:   &eventType,
		Tags:        tags,
		ValueMetric: &metric,
	}

	start := time.Now().UnixNano()

LOOP:
	for {
		select {
		case t := <-tickerChan:
			timestamp = t.UnixNano()
			event.Timestamp = &timestamp
			consumer.events <- event
			eventSent += 1
			if consumer.totalEvents > 0 && eventSent >= consumer.totalEvents {
				break LOOP
			}

			if eventSent%int64(consumer.eps) == 0 {
				duration := time.Now().UnixNano() - start
				fmt.Printf("Generated %d events in %d nano-seconds, actual_eps=%d, required_eps=%d\n",
					eventSent, duration, eventSent*1000000000/duration, consumer.eps)
			}
		case <-consumer.done:
			var done struct{}
			consumer.done <- done
			break LOOP
		}
	}

	duration := time.Now().UnixNano() - start
	fmt.Printf("Down with generation. Generated %d events in %d nano-seconds, actual_eps=%d, required_eps=%d\n",
		eventSent, duration, eventSent*1000000000/duration, consumer.eps)
}
