package testing

import (
	"errors"
	"sync"

	"github.com/cloudfoundry/sonde-go/events"
)

type EventRouterMock struct {
	lock           sync.Mutex
	events         []*events.Envelope
	MockRouteError bool
}

func NewEventRouterMock(mockRouteError bool) *EventRouterMock {
	return &EventRouterMock{MockRouteError: mockRouteError}
}

func (router *EventRouterMock) Route(msg *events.Envelope) error {
	if router.MockRouteError {
		return errors.New("mockup error")
	}
	router.lock.Lock()
	router.events = append(router.events, msg)
	router.lock.Unlock()
	return nil
}

func (router *EventRouterMock) Events() []*events.Envelope {
	var events []*events.Envelope

	router.lock.Lock()
	events = append(events, router.events...)
	router.lock.Unlock()

	return events
}
