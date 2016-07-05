package splunk

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"

	"github.com/cloudfoundry/sonde-go/events"
)

type SplunkEventSerializer struct{}

type SplunkEvent struct {
	Time       string `json:"time,omitempty"`
	Host       string `json:"host,omitempty"`
	Source     string `json:"source,omitempty"`
	SourceType string `json:"sourcetype,omitempty"`
	Index      string `json:"index,omitempty"`

	Event interface{} `json:"event"`
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
		Source: nozzleEvent.GetJob(),
	}
	return splunkEvent
}

type SplunkHttpStartStopMetric struct {
	CommonMetricFields
	StartTimestamp int64    `json:"startTimestamp"`
	StopTimestamp  int64    `json:"stopTimestamp"`
	RequestId      string   `json:"requestId"`
	PeerType       string   `json:"peerType"`
	Method         string   `json:"method"`
	Uri            string   `json:"uri"`
	RemoteAddress  string   `json:"remoteAddress"`
	UserAgent      string   `json:"userAgent"`
	StatusCode     int32    `json:"statusCode"`
	ContentLength  int64    `json:"contentLength"`
	ApplicationId  string   `json:"applicationId"`
	InstanceIndex  int32    `json:"instanceIndex"`
	Forwarded      []string `json:"forwarded"`
}

func (s *SplunkEventSerializer) BuildHttpStartStopEvent(nozzleEvent *events.Envelope) interface{} {
	startStop := nozzleEvent.HttpStartStop

	splunkHttpStartStopMetric := SplunkHttpStartStopMetric{
		StartTimestamp: startStop.GetStartTimestamp(),
		StopTimestamp:  startStop.GetStopTimestamp(),
		RequestId:      uuidToHex(startStop.GetRequestId()),
		PeerType:       startStop.GetPeerType().String(),
		Method:         startStop.GetMethod().String(),
		Uri:            startStop.GetUri(),
		RemoteAddress:  startStop.GetRemoteAddress(),
		UserAgent:      startStop.GetUserAgent(),
		StatusCode:     startStop.GetStatusCode(),
		ContentLength:  startStop.GetContentLength(),
		ApplicationId:  uuidToHex(startStop.GetApplicationId()),
		InstanceIndex:  startStop.GetInstanceIndex(),
		Forwarded:      startStop.GetForwarded(),
	}

	splunkEvent := buildSplunkMetric(nozzleEvent, &splunkHttpStartStopMetric.CommonMetricFields)
	splunkEvent.Event = splunkHttpStartStopMetric
	return splunkEvent
}

type SplunkLogMessageMetric struct {
	CommonMetricFields
	Message        string `json:"logMessage"`
	MessageType    string `json:"MessageType"`
	Timestamp      int64  `json:"timestamp"`
	AppId          string `json:"appId"`
	SourceType     string `json:"sourceType"`
	SourceInstance string `json:"sourceInstance"`
}

func (s *SplunkEventSerializer) BuildLogMessageEvent(nozzleEvent *events.Envelope) interface{} {
	logMessageMetric := nozzleEvent.LogMessage
	splunkLogMessageMetric := SplunkLogMessageMetric{
		Message:        string(logMessageMetric.GetMessage()),
		MessageType:    logMessageMetric.GetMessageType().String(),
		Timestamp:      logMessageMetric.GetTimestamp(),
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

func (s *SplunkEventSerializer) BuildValueMetricEvent(nozzleEvent *events.Envelope) interface{} {
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

func (s *SplunkEventSerializer) BuildCounterEvent(nozzleEvent *events.Envelope) interface{} {
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

func (s *SplunkEventSerializer) BuildErrorEvent(nozzleEvent *events.Envelope) interface{} {
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

func (s *SplunkEventSerializer) BuildContainerEvent(nozzleEvent *events.Envelope) interface{} {
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

const dashByte byte = '-'

func uuidToHex(uuid *events.UUID) string {
	if uuid == nil {
		return ""
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 16))
	binary.Write(buffer, binary.LittleEndian, uuid.Low)
	binary.Write(buffer, binary.LittleEndian, uuid.High)
	bufferBytes := buffer.Bytes()

	hexBuffer := make([]byte, 36)
	hex.Encode(hexBuffer[0:8], bufferBytes[0:4])
	hexBuffer[8] = dashByte
	hex.Encode(hexBuffer[9:13], bufferBytes[4:6])
	hexBuffer[13] = dashByte
	hex.Encode(hexBuffer[14:18], bufferBytes[6:8])
	hexBuffer[18] = dashByte
	hex.Encode(hexBuffer[19:23], bufferBytes[8:10])
	hexBuffer[23] = dashByte
	hex.Encode(hexBuffer[24:], bufferBytes[10:])

	return string(hexBuffer)
}
