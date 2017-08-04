package testing

type MemorySink struct {
	Events   []map[string]interface{}
	Messages []string
}

func NewMemorySink() *MemorySink {
	return &MemorySink{
		Events:   []map[string]interface{}{},
		Messages: []string{},
	}
}

func (l *MemorySink) Open() error {
	return nil
}

func (l *MemorySink) Close() error {
	return nil
}

func (l *MemorySink) Write(fields map[string]interface{}, msg string) error {
	l.Events = append(l.Events, fields)
	l.Messages = append(l.Messages, msg)
	return nil
}
