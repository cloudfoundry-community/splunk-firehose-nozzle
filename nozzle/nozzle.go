package nozzle

import (
	"fmt"
	"math"

	"github.com/cloudfoundry/sonde-go/events"
)

type Nozzle interface {
	Run() error
}

type SplunkNozzle struct {
	splunkClient SplunkClient
	events       <-chan *events.Envelope
	errors       <-chan error
}

func NewSplunkForwarder(splunkClient SplunkClient, events <-chan *events.Envelope, errors <-chan error) Nozzle {
	return &SplunkNozzle{
		splunkClient: splunkClient,
		events:       events,
		errors:       errors,
	}
}

func (s *SplunkNozzle) Run() error {
	go func() {
		for event := range s.events {
			s.handleEvent(event)
		}
	}()

	return <-s.errors
}

func (s *SplunkNozzle) handleEvent(event *events.Envelope) {
	var splunkEvent *SplunkEvent = nil

	switch *event.EventType {
	case events.Envelope_HttpStart:
	case events.Envelope_HttpStop:
	case events.Envelope_HttpStartStop:
	case events.Envelope_LogMessage:
		splunkEvent = BuildLogMessageMetric(event)
	case events.Envelope_ValueMetric:
		splunkEvent = BuildValueMetric(event)
	case events.Envelope_CounterEvent:
		splunkEvent = BuildCounterEventMetric(event)
	case events.Envelope_Error:
		splunkEvent = BuildErrorMetric(event)
	case events.Envelope_ContainerMetric:
		splunkEvent = BuildContainerMetric(event)
	}

	if splunkEvent != nil {
		s.splunkClient.Post(splunkEvent)
	}
}

type CommonMetricFields struct {
	Deployment string `json:"deployment"`
	Index      string `json:"index"`
	EventType  string `json:"eventType"`
}

func buildSplunkMetric(nozzleEvent *events.Envelope, shared *CommonMetricFields) *SplunkEvent {
	shared.Deployment = nozzleEvent.GetDeployment()
	shared.Index = nozzleEvent.GetIndex()
	shared.EventType = nozzleEvent.GetEventType().String()

	splunkEvent := &SplunkEvent{
		Time:   nanoSecondsToSeconds(nozzleEvent.GetTimestamp()),
		Host:   nozzleEvent.GetIp(),
		Source: nozzleEvent.GetJob(), //todo: consider app vs cf once understand full metric set
	}
	return splunkEvent
}

type SplunkLogMessageMetric struct {
	CommonMetricFields
	Message        string `json:"logMessage"`
	MessageType    string `json:"MessageType"`
	Timestamp      string `json:"timestampe"`
	AppId          string `json:"appId"`
	SourceType     string `json:"sourceType"`
	SourceInstance string `json:"sourceInstance"`
}

func BuildLogMessageMetric(nozzleEvent *events.Envelope) *SplunkEvent {
	logMessageMetric := nozzleEvent.LogMessage
	splunkLogMessageMetric := SplunkLogMessageMetric{
		Message:        string(logMessageMetric.GetMessage()),
		MessageType:    logMessageMetric.GetMessageType().String(),
		Timestamp:      nanoSecondsToSeconds(logMessageMetric.GetTimestamp()),
		AppId:          logMessageMetric.GetAppId(),
		SourceType:     logMessageMetric.GetSourceType(),
		SourceInstance: logMessageMetric.GetSourceInstance(),
	}

	splunkEvent := buildSplunkMetric(nozzleEvent, &splunkLogMessageMetric.CommonMetricFields)
	splunkEvent.Event = splunkLogMessageMetric
	return splunkEvent
}

type SplunkValueMetric struct {
	CommonMetricFields
	Name  string  `json:"name"`
	Value float64 `json:"value"`
	Unit  string  `json:"unit"`
}

func BuildValueMetric(nozzleEvent *events.Envelope) *SplunkEvent {
	valueMetric := nozzleEvent.ValueMetric
	splunkValueMetric := SplunkValueMetric{
		Name:  valueMetric.GetName(),
		Value: valueMetric.GetValue(),
		Unit:  valueMetric.GetUnit(),
	}

	splunkEvent := buildSplunkMetric(nozzleEvent, &splunkValueMetric.CommonMetricFields)
	splunkEvent.Event = splunkValueMetric
	return splunkEvent
}

type SplunkCounterEventMetric struct {
	CommonMetricFields
	Name  string `json:"name"`
	Delta uint64 `json:"delta"`
	Total uint64 `json:"total"`
}

func BuildCounterEventMetric(nozzleEvent *events.Envelope) *SplunkEvent {
	counterEvent := nozzleEvent.GetCounterEvent()
	splunkCounterEventMetric := SplunkCounterEventMetric{
		Name:  counterEvent.GetName(),
		Delta: counterEvent.GetDelta(),
		Total: counterEvent.GetTotal(),
	}

	splunkEvent := buildSplunkMetric(nozzleEvent, &splunkCounterEventMetric.CommonMetricFields)
	splunkEvent.Event = splunkCounterEventMetric
	return splunkEvent
}

type SplunkErrorMetric struct {
	CommonMetricFields
	Source  string `json:"source"`
	Code    int32  `json:"code"`
	Message string `json:"message"`
}

func BuildErrorMetric(nozzleEvent *events.Envelope) *SplunkEvent {
	errorMetric := nozzleEvent.Error
	splunkErrorMetric := SplunkErrorMetric{
		Source:  errorMetric.GetSource(),
		Code:    errorMetric.GetCode(),
		Message: errorMetric.GetMessage(),
	}

	splunkEvent := buildSplunkMetric(nozzleEvent, &splunkErrorMetric.CommonMetricFields)
	splunkEvent.Event = splunkErrorMetric
	return splunkEvent
}

type SplunkContainerMetric struct {
	CommonMetricFields
	ApplicationId string  `json:"applicationId"`
	InstanceIndex int32   `json:"instanceIndex"`
	CpuPercentage float64 `json:"cpuPercentage"`
	MemoryBytes   uint64  `json:"memoryBytes"`
	DiskBytes     uint64  `json:"diskBytes"`
}

func BuildContainerMetric(nozzleEvent *events.Envelope) *SplunkEvent {
	containerMetric := nozzleEvent.GetContainerMetric()
	splunkContainerMetric := SplunkContainerMetric{
		ApplicationId: containerMetric.GetApplicationId(),
		InstanceIndex: containerMetric.GetInstanceIndex(),
		CpuPercentage: containerMetric.GetCpuPercentage(),
		MemoryBytes:   containerMetric.GetMemoryBytes(),
		DiskBytes:     containerMetric.GetDiskBytes(),
	}

	splunkEvent := buildSplunkMetric(nozzleEvent, &splunkContainerMetric.CommonMetricFields)
	splunkEvent.Event = splunkContainerMetric
	return splunkEvent
}

func nanoSecondsToSeconds(nanoseconds int64) string {
	seconds := float64(nanoseconds) * math.Pow(1000, -3)
	return fmt.Sprintf("%.3f", seconds)
}
