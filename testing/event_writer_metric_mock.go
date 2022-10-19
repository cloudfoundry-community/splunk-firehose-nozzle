package testing

import (
	"errors"
	"sync"
	"time"
)

type EventWriterMetricMock struct {
	lock           sync.Mutex
	Block          bool
	CapturedEvents []map[string]interface{}
	PostBatchFn    func(events []map[string]interface{}) error
	ReturnErr      bool
}

func (m *EventWriterMetricMock) Write(events []map[string]interface{}) (error, uint64) {
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
		m.CapturedEvents = append(m.CapturedEvents, events[0])
		m.lock.Unlock()
	}
	return nil, uint64(len(events))
}
