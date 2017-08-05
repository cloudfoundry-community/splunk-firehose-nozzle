package testing

import (
	"fmt"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gorilla/websocket"
)

type MemoryEventSourceMock struct {
	events      chan *events.Envelope
	errors      chan error
	eps         int
	totalEvents int64
	done        chan struct{}
}

const (
	maxEPS = 500000
)

func biggerThanMaxEPS(eps int) bool {
	return eps <= 0 || eps > maxEPS
}

func NewMemoryEventSourceMock(eps int, totalEvents int64, errCode int) *MemoryEventSourceMock {
	e := &MemoryEventSourceMock{
		events:      make(chan *events.Envelope, 1000000),
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

		e.errors <- err
	}

	go e.publishEvents()
	return e
}

func (e *MemoryEventSourceMock) Open() error {
	return nil
}

func (e *MemoryEventSourceMock) Read() (<-chan *events.Envelope, <-chan error) {
	return e.events, e.errors
}

func (e *MemoryEventSourceMock) Close() error {
	var done struct{}
	e.done <- done
	<-e.done
	return nil
}

func (e *MemoryEventSourceMock) produce(numOfEvents int64) {
	event := newEvent()
	for i := int64(0); i < numOfEvents; i++ {
		t := time.Now().UnixNano()
		event.Timestamp = &t
		e.events <- event
	}
}

func (e *MemoryEventSourceMock) publishEvents() {
	if biggerThanMaxEPS(e.eps) {
		e.publishEventsAsFastAsPossible()
		return
	}

	// 5 seconds as a window
	windowEvents := int64(e.eps * 5)
	windowDuration := time.Duration(5) * time.Second

	eventSent := int64(0)
	start := time.Now().UnixNano()

LOOP:
	for {
		produceStart := time.Now().UnixNano()
		if e.totalEvents > 0 && eventSent+windowEvents > e.totalEvents {
			windowEvents = eventSent + windowEvents - e.totalEvents
		}
		e.produce(windowEvents)
		eventSent += windowEvents
		duration := time.Duration(time.Now().UnixNano() - produceStart)
		if duration < windowDuration {
			fmt.Printf("Too fast, sleep %d nano-seconds\n", int64(windowDuration-duration))
			time.Sleep(windowDuration - duration)
		} else {
			fmt.Printf("Too slow, over committed=%d nano-seconds\n", int64(duration-windowDuration))
		}

		if eventSent%int64(e.eps) == 0 {
			duration := time.Now().UnixNano() - start
			fmt.Printf("Generated %d events in %d nano-seconds, actual_eps=%d, required_eps=%d\n",
				eventSent, duration, eventSent*1000000000/duration, e.eps)
		}

		if e.totalEvents > 0 && eventSent >= e.totalEvents {
			break LOOP
		}

		select {
		case <-e.done:
			var done struct{}
			e.done <- done
			break LOOP
		default:
		}
	}

	duration := time.Now().UnixNano() - start
	fmt.Printf("Done with generation. Generated %d events in %d nano-seconds, actual_eps=%d, required_eps=%d\n",
		eventSent, duration, eventSent*1000000000/duration, e.eps)
}

func (e *MemoryEventSourceMock) publishEventsAsFastAsPossible() {
	eventSent := int64(0)
	event := newEvent()
	start := time.Now().UnixNano()

LOOP:
	for {
		timestamp := time.Now().UnixNano()
		event.Timestamp = &timestamp

		select {
		case e.events <- event:
			eventSent += 1
			if e.totalEvents > 0 && eventSent >= e.totalEvents {
				break LOOP
			}

			if eventSent%maxEPS == 0 {
				duration := time.Now().UnixNano() - start
				fmt.Printf("Generated %d events in %d nano-seconds, actual_eps=%d, required_eps=%d\n",
					eventSent, duration, eventSent*1000000000/duration, e.eps)
			}
		case <-e.done:
			var done struct{}
			e.done <- done
			break LOOP
		}
	}

	duration := time.Now().UnixNano() - start
	fmt.Printf("Done with generation. Generated %d events in %d nano-seconds, actual_eps=%d, required_eps=%d\n",
		eventSent, duration, eventSent*1000000000/duration, e.eps)
}

func newEvent() *events.Envelope {
	var (
		origin     = "DopplerServer"
		deployment = "cf"
		job        = "doppler"
		index      = "5a634d0b-bbc5-47c4-9450-a43f44a7fd30"
		ip         = "192.168.16.26"

		eventType    = events.Envelope_ValueMetric
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

	t := time.Now().UnixNano()
	event := &events.Envelope{
		Origin:      &origin,
		Deployment:  &deployment,
		Job:         &job,
		Index:       &index,
		Ip:          &ip,
		EventType:   &eventType,
		Tags:        tags,
		ValueMetric: &metric,
		Timestamp:   &t,
	}
	return event
}
