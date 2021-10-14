package events_test

import (
	"testing"

	. "github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Events Suite")
}

func NewLogMessage() *Envelope {
	var eventType Envelope_EventType = 5
	var messageType LogMessage_MessageType = 1
	var posixStart int64 = 1
	var origin string = "yomomma__0"
	var sourceType string = "Kehe"
	var logMsg string = "Help, I'm a rock! Help, I'm a rock! Help, I'm a cop! Help, I'm a cop!"
	var sourceInstance string = ">9000"
	var appID string = "eea38ba5-53a5-4173-9617-b442d35ec2fd"
	var tags map[string]string = map[string]string{"key": "value"}

	logMessage := LogMessage{
		Message:        []byte(logMsg),
		AppId:          &appID,
		Timestamp:      &posixStart,
		SourceType:     &sourceType,
		MessageType:    &messageType,
		SourceInstance: &sourceInstance,
	}

	envelope := &Envelope{
		EventType:  &eventType,
		Origin:     &origin,
		LogMessage: &logMessage,
		Tags:       tags,
	}
	return envelope
}

var (
	origin                             = "yomomma__0"
	eventType   Envelope_EventType     = 5
	messageType LogMessage_MessageType = 1

	timestamp  int64 = 1
	deployment       = "deployment"
	job              = "job"
	index            = "index"
	ip               = "127.0.0.1"
	tags             = map[string]string{"tag": "value"}

	// `f47ac10b-58cc-4372-a567-0e02b2c3d479` should be encoded as `UUID{ low: 0x7243cc580bc17af4, high: 0x79d4c3b2020e67a5 }`
	low  uint64 = 0x7243cc580bc17af4
	high uint64 = 0x79d4c3b2020e67a5
	uuid        = UUID{
		Low:  &low,
		High: &high,
	}
	uuidStr = "f47ac10b-58cc-4372-a567-0e02b2c3d479"

	requestId       = uuid
	peerType        = PeerType_Client
	peerTypeStr     = "Client"
	method          = Method_GET
	methodStr       = "GET"
	uri             = "localhost"
	remoteAddr      = "localhost"
	userAgent       = "agent"
	parentRequestId = uuid
	appId           = uuid
	instanceId      = "instanceid"

	instanceIdx   int32 = 0
	statusCode    int32 = 0
	contentLength int64 = 0

	name  = "cpu"
	value = 10.0
	unit  = "percentage"

	delta uint64 = 10
	total uint64 = 100

	source        = "source"
	code    int32 = -1
	message       = "hello, exception"

	cpuPercentage           = 10.0
	memoryBytes      uint64 = 1024
	diskBytes        uint64 = 1024
	memoryBytesQuota uint64 = 10240
	diskBytesQuota   uint64 = 10240

	event = &Envelope{
		Origin:     &origin,
		EventType:  &eventType,
		Timestamp:  &timestamp,
		Deployment: &deployment,
		Job:        &job,
		Index:      &index,
		Ip:         &ip,
		Tags:       tags,
	}
)

func NewHttpStart() *Envelope {
	httpStart := &HttpStart{
		Timestamp:       &timestamp,
		RequestId:       &requestId,
		PeerType:        &peerType,
		Method:          &method,
		Uri:             &uri,
		RemoteAddress:   &remoteAddr,
		UserAgent:       &userAgent,
		ParentRequestId: &parentRequestId,
		ApplicationId:   &appId,
		InstanceIndex:   &instanceIdx,
		InstanceId:      &instanceId,
	}
	event.HttpStart = httpStart

	return event
}

func NewHttpStop() *Envelope {
	httpStop := &HttpStop{
		Timestamp:     &timestamp,
		Uri:           &uri,
		RequestId:     &requestId,
		PeerType:      &peerType,
		StatusCode:    &statusCode,
		ContentLength: &contentLength,
		ApplicationId: &appId,
	}
	event.HttpStop = httpStop

	return event
}

func NewHttpStartStop() *Envelope {
	httpStartStop := &HttpStartStop{
		StartTimestamp: &timestamp,
		StopTimestamp:  &timestamp,
		RequestId:      &requestId,
		PeerType:       &peerType,
		Method:         &method,
		Uri:            &uri,
		RemoteAddress:  &remoteAddr,
		UserAgent:      &userAgent,
		StatusCode:     &statusCode,
		ContentLength:  &contentLength,
		ApplicationId:  &appId,
		InstanceIndex:  &instanceIdx,
		InstanceId:     &instanceId,
	}
	event.HttpStartStop = httpStartStop

	return event
}

func NewValueMetric() *Envelope {
	valueMetric := &ValueMetric{
		Name:  &name,
		Value: &value,
		Unit:  &unit,
	}
	event.ValueMetric = valueMetric

	return event
}

func NewCounterEvent() *Envelope {
	counterEvent := &CounterEvent{
		Name:  &name,
		Delta: &delta,
		Total: &total,
	}
	event.CounterEvent = counterEvent

	return event
}

func NewErrorEvent() *Envelope {
	err := &Error{
		Source:  &source,
		Code:    &code,
		Message: &message,
	}
	event.Error = err

	return event
}

func NewContainerMetric() *Envelope {
	metric := &ContainerMetric{
		ApplicationId:    &uuidStr,
		InstanceIndex:    &instanceIdx,
		CpuPercentage:    &cpuPercentage,
		MemoryBytes:      &memoryBytes,
		DiskBytes:        &diskBytes,
		MemoryBytesQuota: &memoryBytesQuota,
		DiskBytesQuota:   &diskBytesQuota,
	}
	event.ContainerMetric = metric

	return event
}
