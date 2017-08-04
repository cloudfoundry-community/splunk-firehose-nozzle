package drain

type LoggingMemory struct {
	Events   []map[string]interface{}
	Messages []string
}

func NewLoggingMemory() *LoggingMemory {
	return &LoggingMemory{
		Events:   []map[string]interface{}{},
		Messages: []string{},
	}
}

func (l *LoggingMemory) Open() error {
	return nil
}

func (l *LoggingMemory) Close() error {
	return nil
}

func (l *LoggingMemory) Log(fields map[string]interface{}, msg string) error {
	l.Events = append(l.Events, fields)
	l.Messages = append(l.Messages, msg)
	return nil
}
