package testing

import (
	"errors"
	"sync"
	"time"
)

type EventWriterMock struct {
	lock           sync.Mutex
	Block          bool
	capturedEvents []map[string]interface{}
	PostBatchFn    func(events []map[string]interface{}) error
	ReturnErr      bool
}

func (m *EventWriterMock) Write(events []map[string]interface{}) (error, uint64) {
	if m.Block {
		time.Sleep(time.Millisecond * 100)
	}
	if m.ReturnErr {
		return errors.New("mockup error"), 0
	}

	if m.PostBatchFn != nil {
		return m.PostBatchFn(events), 0
	} else {
		m.lock.Lock()
		m.capturedEvents = append(m.capturedEvents, events...)
		m.lock.Unlock()
	}
	return nil, uint64(len(m.capturedEvents))
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
