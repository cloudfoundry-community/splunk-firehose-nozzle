package testing

import (
	"errors"
	"sync"
)

type EventWriterMock struct {
	lock           sync.Mutex
	capturedEvents []map[string]interface{}
	PostBatchFn    func(events []map[string]interface{}) error
	ReturnErr      bool
}

func (m *EventWriterMock) Write(events []map[string]interface{}) error {
	if m.ReturnErr {
		return errors.New("mockup error")
	}

	if m.PostBatchFn != nil {
		return m.PostBatchFn(events)
	} else {
		m.lock.Lock()
		m.capturedEvents = append(m.capturedEvents, events...)
		m.lock.Unlock()
	}
	return nil
}

func (m *EventWriterMock) CapturedEvents() []map[string]interface{} {
	m.lock.Lock()
	var events []map[string]interface{}
	for _, event := range m.capturedEvents {
		events = append(events, event)
	}
	m.lock.Unlock()

	return events
}
