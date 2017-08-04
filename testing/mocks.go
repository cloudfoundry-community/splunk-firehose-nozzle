package testing

import (
	"sync"
)

type MockEventWriter struct {
	lock           sync.Mutex
	capturedEvents []map[string]interface{}
	PostBatchFn    func(events []map[string]interface{}) error
}

func (m *MockEventWriter) Write(events []map[string]interface{}) error {
	if m.PostBatchFn != nil {
		return m.PostBatchFn(events)
	} else {
		m.lock.Lock()
		m.capturedEvents = append(m.capturedEvents, events...)
		m.lock.Unlock()
	}
	return nil
}

func (m *MockEventWriter) CapturedEvents() []map[string]interface{} {
	m.lock.Lock()
	var events []map[string]interface{}
	for _, event := range m.capturedEvents {
		events = append(events, event)
	}
	m.lock.Unlock()

	return events
}

type MockTokenGetter struct {
	GetTokenFn func() (string, error)
}

func (m *MockTokenGetter) GetToken() (string, error) {
	if m.GetTokenFn != nil {
		return m.GetTokenFn()
	} else {
		return "", nil
	}
}
