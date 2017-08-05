package testing

type MemorySinkMock struct {
	Events   []map[string]interface{}
	Messages []string
}

func NewMemorySinkMock() *MemorySinkMock {
	return &MemorySinkMock{
		Events:   []map[string]interface{}{},
		Messages: []string{},
	}
}

func (l *MemorySinkMock) Open() error {
	return nil
}

func (l *MemorySinkMock) Close() error {
	return nil
}

func (l *MemorySinkMock) Write(fields map[string]interface{}, msg string) error {
	l.Events = append(l.Events, fields)
	l.Messages = append(l.Messages, msg)
	return nil
}
