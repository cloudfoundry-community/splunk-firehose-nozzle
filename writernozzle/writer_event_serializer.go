package writernozzle

import (
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
)

type WriterEventSerializer struct{}

func NewWriterEventSerializer() *WriterEventSerializer {
	return &WriterEventSerializer{}
}

func (w *WriterEventSerializer) BuildHttpStartStopEvent(event *events.Envelope) interface{} {
	return []byte(fmt.Sprintf("%+v\n", event))
}

func (w *WriterEventSerializer) BuildLogMessageEvent(event *events.Envelope) interface{} {
	return []byte(fmt.Sprintf("%+v\n", event))
}

func (w *WriterEventSerializer) BuildValueMetricEvent(event *events.Envelope) interface{} {
	return []byte(fmt.Sprintf("%+v\n", event))
}

func (w *WriterEventSerializer) BuildCounterEvent(event *events.Envelope) interface{} {
	return []byte(fmt.Sprintf("%+v\n", event))
}

func (w *WriterEventSerializer) BuildErrorEvent(event *events.Envelope) interface{} {
	return []byte(fmt.Sprintf("%+v\n", event))
}

func (w *WriterEventSerializer) BuildContainerEvent(event *events.Envelope) interface{} {
	return []byte(fmt.Sprintf("%+v\n", event))
}
