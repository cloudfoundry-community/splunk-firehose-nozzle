package testing

import (
	"errors"

	"github.com/cloudfoundry/sonde-go/events"
)

type MemorySinkMock struct {
	Events    []*events.Envelope
	ReturnErr bool
}

func NewMemorySinkMock() *MemorySinkMock {
	return &MemorySinkMock{
		Events: []*events.Envelope{},
	}
}

func (l *MemorySinkMock) Open() error {
	return nil
}

func (l *MemorySinkMock) Close() error {
	return nil
}

func (l *MemorySinkMock) Write(fields *events.Envelope) error {
	if l.ReturnErr {
		return errors.New("mockup error")
	}

	l.Events = append(l.Events, fields)
	// l.Messages = append(l.Messages, msg)
	return nil
}
