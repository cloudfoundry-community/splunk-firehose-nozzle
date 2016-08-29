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

func (l *LoggingMemory) Connect() bool {
	return true
}

func (l *LoggingMemory) ShipEvents(fields map[string]interface{}, msg string) {
	l.Events = append(l.Events, fields)
	l.Messages = append(l.Messages, msg)
}
