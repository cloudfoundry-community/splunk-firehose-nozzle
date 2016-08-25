package splunk_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/splunk"

	"encoding/json"
	"github.com/cloudfoundry/sonde-go/events"
)

var _ = Describe("SplunkEventSerializer", func() {

	var (
		splunkEventSerializer *SplunkEventSerializer

		origin        string
		deployment    string
		job           string
		jobIndex      string
		ip            string
		timestampNano int64
		envelope      *events.Envelope
		eventType     events.Envelope_EventType

		metric *SplunkEvent
	)

	BeforeEach(func() {
		timestampNano = 1467040874046121775
		deployment = "cf-warden"
		jobIndex = "85c9ff80-e99b-470b-a194-b397a6e73913"
		ip = "10.244.0.22"
		envelope = &events.Envelope{
			Origin:     &origin,
			EventType:  &eventType,
			Timestamp:  &timestampNano,
			Deployment: &deployment,
			Job:        &job,
			Index:      &jobIndex,
			Ip:         &ip,
		}

		splunkEventSerializer = &SplunkEventSerializer{}
	})

	It("properly serializes to splunk json", func() {
		var envelopeValueMetric *events.ValueMetric

		name := "ms_since_last_registry_update"
		value := 1581.0
		unit := "ms"
		envelopeValueMetric = &events.ValueMetric{
			Name:  &name,
			Value: &value,
			Unit:  &unit,
		}

		job = "router_z1"
		origin = "MetronAgent"
		eventType = events.Envelope_ValueMetric
		envelope.ValueMetric = envelopeValueMetric

		metric = splunkEventSerializer.BuildValueMetricEvent(envelope).(*SplunkEvent)

		marshalled, err := json.MarshalIndent(metric, "", "    ")
		Expect(err).To(BeNil())

		Expect(string(marshalled)).To(Equal(`{
    "time": "1467040874.046",
    "host": "10.244.0.22",
    "source": "router_z1",
    "sourcetype": "cf:valuemetric",
    "event": {
        "deployment": "cf-warden",
        "jobIndex": "85c9ff80-e99b-470b-a194-b397a6e73913",
        "eventType": "ValueMetric",
        "origin": "MetronAgent",
        "name": "ms_since_last_registry_update",
        "value": 1581,
        "unit": "ms"
    }
}`))
	})

	Context("envelope HttpStartStop", func() {
		var envelopeHttpStartStop *events.HttpStartStop
		var startTimestamp, stopTimestamp int64
		var requestId events.UUID
		var requestIdHex, applicationIdHex string
		var peerType events.PeerType
		var method events.Method
		var uri, remoteAddress, userAgent string
		var statusCode int32
		var contentLength int64
		var applicationId events.UUID
		var instanceIndex int32
		var instanceId string
		var forwarded []string

		BeforeEach(func() {
			startTimestamp = 1467143062034348090
			stopTimestamp = 1467143062042890400
			peerType = events.PeerType_Server
			method = events.Method_GET
			uri = "http://app-node-express.bosh-lite.com/"
			remoteAddress = "10.244.0.34:45334"
			userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36"
			statusCode = 200
			contentLength = 23
			instanceIndex = 1
			instanceId = "055a847afbb146f78fdcebf6b2a0067bef2394a07fd34d06a3c3e0811aa966ee"
			forwarded = []string{"hello"}

			requestIdLow := uint64(17459518436806699697)
			requestIdHigh := uint64(17377260946761993045)
			requestId = events.UUID{
				Low:  &requestIdLow,
				High: &requestIdHigh,
			}
			requestIdHex = "b12a3f87-83ab-4cf2-554b-042dc36e28f1"

			applicationIdLow := uint64(10539615360601842564)
			applicationIdHigh := uint64(3160954123591206558)
			applicationId = events.UUID{
				Low:  &applicationIdLow,
				High: &applicationIdHigh,
			}
			applicationIdHex = "8463ec45-543c-4492-9ec6-f52707f7dd2b"

			envelopeHttpStartStop = &events.HttpStartStop{
				StartTimestamp: &startTimestamp,
				StopTimestamp:  &stopTimestamp,
				RequestId:      &requestId,
				PeerType:       &peerType,
				Method:         &method,
				Uri:            &uri,
				RemoteAddress:  &remoteAddress,
				UserAgent:      &userAgent,
				StatusCode:     &statusCode,
				ContentLength:  &contentLength,
				ApplicationId:  &applicationId,
				InstanceIndex:  &instanceIndex,
				InstanceId:     &instanceId,
				Forwarded:      forwarded,
			}

			job = "runner_z1"
			origin = "gorouter"
			eventType = events.Envelope_HttpStartStop
			envelope.HttpStartStop = envelopeHttpStartStop

			metric = splunkEventSerializer.BuildHttpStartStopEvent(envelope).(*SplunkEvent)
		})

		It("metadata", func() {
			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			//Expect(metric.Index).To(Equal(index))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
			Expect(metric.SourceType).To(Equal("cf:httpstartstop"))
		})

		It("common components", func() {
			event := metric.Event.(SplunkHttpStartStopMetric)

			Expect(event.Deployment).To(Equal("cf-warden"))
			Expect(event.JobIndex).To(Equal("85c9ff80-e99b-470b-a194-b397a6e73913"))
			Expect(event.EventType).To(Equal("HttpStartStop"))
			Expect(event.Origin).To(Equal("gorouter"))
		})

		It("log message specific data", func() {
			event := metric.Event.(SplunkHttpStartStopMetric)

			Expect(event.StartTimestamp).To(Equal(startTimestamp))
			Expect(event.StopTimestamp).To(Equal(stopTimestamp))
			Expect(event.RequestId).To(Equal(requestIdHex))
			Expect(event.PeerType).To(Equal("Server"))
			Expect(event.Method).To(Equal("GET"))
			Expect(event.Uri).To(Equal(uri))
			Expect(event.RemoteAddress).To(Equal(remoteAddress))
			Expect(event.UserAgent).To(Equal(userAgent))
			Expect(event.StatusCode).To(Equal(statusCode))
			Expect(event.ContentLength).To(Equal(contentLength))
			Expect(event.ApplicationId).To(Equal(applicationIdHex))
			Expect(event.InstanceIndex).To(Equal(instanceIndex))
			Expect(event.Forwarded).To(Equal(forwarded))
		})
	})

	Context("envelope LogMessage", func() {
		var message []byte
		var messageType events.LogMessage_MessageType
		var timestamp int64
		var appId, sourceType, sourceInstance string
		var envelopeLogMessage *events.LogMessage

		BeforeEach(func() {
			message = []byte("App debug log message")
			messageType = events.LogMessage_OUT
			timestamp = 1467128185055072010
			appId = "8463ec45-543c-4492-9ec6-f52707f7dd2b"
			sourceType = "App"
			sourceInstance = "0"
			envelopeLogMessage = &events.LogMessage{
				Message:        message,
				MessageType:    &messageType,
				Timestamp:      &timestamp,
				AppId:          &appId,
				SourceType:     &sourceType,
				SourceInstance: &sourceInstance,
			}

			job = "runner_z1"
			origin = "dea_logging_agent"
			eventType = events.Envelope_LogMessage
			envelope.LogMessage = envelopeLogMessage

			metric = splunkEventSerializer.BuildLogMessageEvent(envelope).(*SplunkEvent)
		})

		BeforeEach(func() {

		})

		It("metadata", func() {
			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
			Expect(metric.SourceType).To(Equal("cf:logmessage"))
		})

		It("common components", func() {
			event := metric.Event.(SplunkLogMessageMetric)

			Expect(event.Deployment).To(Equal("cf-warden"))
			Expect(event.JobIndex).To(Equal("85c9ff80-e99b-470b-a194-b397a6e73913"))
			Expect(event.EventType).To(Equal("LogMessage"))
			Expect(event.Origin).To(Equal("dea_logging_agent"))

		})

		It("log message specific data", func() {
			event := metric.Event.(SplunkLogMessageMetric)

			Expect(event.Message).To(Equal(string(message)))
			Expect(event.MessageType).To(Equal("OUT"))
			Expect(event.Timestamp).To(Equal(timestamp))
			Expect(event.AppId).To(Equal(appId))
			Expect(event.SourceType).To(Equal(sourceType))
			Expect(event.SourceInstance).To(Equal(sourceInstance))
		})
	})

	Context("envelope ValueMetric", func() {
		var name, unit string
		var value float64
		var envelopeValueMetric *events.ValueMetric

		BeforeEach(func() {
			name = "ms_since_last_registry_update"
			value = 1581.0
			unit = "ms"
			envelopeValueMetric = &events.ValueMetric{
				Name:  &name,
				Value: &value,
				Unit:  &unit,
			}

			job = "router_z1"
			origin = "MetronAgent"
			eventType = events.Envelope_ValueMetric
			envelope.ValueMetric = envelopeValueMetric

			metric = splunkEventSerializer.BuildValueMetricEvent(envelope).(*SplunkEvent)
		})

		It("metadata", func() {
			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
			Expect(metric.SourceType).To(Equal("cf:valuemetric"))
		})

		It("common components", func() {
			event := metric.Event.(SplunkValueMetric)

			Expect(event.Deployment).To(Equal("cf-warden"))
			Expect(event.JobIndex).To(Equal("85c9ff80-e99b-470b-a194-b397a6e73913"))
			Expect(event.EventType).To(Equal("ValueMetric"))
			Expect(event.Origin).To(Equal("MetronAgent"))

		})

		It("metric specific data", func() {
			event := metric.Event.(SplunkValueMetric)

			Expect(event.Name).To(Equal(name))
			Expect(event.Value).To(Equal(value))
			Expect(event.Unit).To(Equal(unit))
		})
	})

	Context("envelope CounterEvent", func() {
		var name string
		var delta, total uint64
		var counterEvent *events.CounterEvent

		BeforeEach(func() {
			name = "registry_message.uaa"
			delta = 1
			total = 8196
			counterEvent = &events.CounterEvent{
				Name:  &name,
				Delta: &delta,
				Total: &total,
			}

			job = "router_z1"
			origin = "gorouter"
			eventType = events.Envelope_CounterEvent
			envelope.CounterEvent = counterEvent

			metric = splunkEventSerializer.BuildCounterEvent(envelope).(*SplunkEvent)
		})

		It("metadata", func() {
			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
			Expect(metric.SourceType).To(Equal("cf:counterevent"))
		})

		It("common components", func() {
			event := metric.Event.(SplunkCounterEventMetric)

			Expect(event.Deployment).To(Equal("cf-warden"))
			Expect(event.JobIndex).To(Equal("85c9ff80-e99b-470b-a194-b397a6e73913"))
			Expect(event.EventType).To(Equal("CounterEvent"))
			Expect(event.Origin).To(Equal("gorouter"))

		})

		It("metric specific data", func() {
			event := metric.Event.(SplunkCounterEventMetric)

			Expect(event.Name).To(Equal(name))
			Expect(event.Delta).To(Equal(delta))
			Expect(event.Total).To(Equal(total))
		})
	})

	Context("envelope Error", func() {
		var source, message string
		var code int32
		var envelopeError *events.Error

		BeforeEach(func() {
			source = "some_source"
			message = "something failed"
			code = 42
			envelopeError = &events.Error{
				Source:  &source,
				Code:    &code,
				Message: &message,
			}

			job = "router_z1"
			origin = "Unknown"
			eventType = events.Envelope_Error
			envelope.Error = envelopeError

			metric = splunkEventSerializer.BuildErrorEvent(envelope).(*SplunkEvent)
		})

		It("metadata", func() {
			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
			Expect(metric.SourceType).To(Equal("cf:error"))
		})

		It("common components", func() {
			event := metric.Event.(SplunkErrorMetric)

			Expect(event.Deployment).To(Equal("cf-warden"))
			Expect(event.JobIndex).To(Equal("85c9ff80-e99b-470b-a194-b397a6e73913"))
			Expect(event.EventType).To(Equal("Error"))
			Expect(event.Origin).To(Equal("Unknown"))

		})

		It("metric specific data", func() {
			event := metric.Event.(SplunkErrorMetric)

			Expect(event.Source).To(Equal(source))
			Expect(event.Code).To(Equal(code))
			Expect(event.Message).To(Equal(message))
		})
	})

	Context("envelope ContainerMetric", func() {
		var applicationId string
		var instanceIndex int32
		var cpuPercentage float64
		var memoryBytes, diskBytes uint64
		var containerMetric *events.ContainerMetric

		BeforeEach(func() {
			applicationId = "8463ec45-543c-4492-9ec6-f52707f7dd2b"
			instanceIndex = 1
			cpuPercentage = 1.0916583859477904
			memoryBytes = 30011392
			diskBytes = 15005696
			containerMetric = &events.ContainerMetric{
				ApplicationId: &applicationId,
				InstanceIndex: &instanceIndex,
				CpuPercentage: &cpuPercentage,
				MemoryBytes:   &memoryBytes,
				DiskBytes:     &diskBytes,
			}

			job = "runner_z1"
			origin = "DEA"
			eventType = events.Envelope_ContainerMetric
			envelope.ContainerMetric = containerMetric

			metric = splunkEventSerializer.BuildContainerEvent(envelope).(*SplunkEvent)
		})

		It("metadata", func() {
			eventTimeSeconds := "1467040874.046"
			Expect(metric.Time).To(Equal(eventTimeSeconds))
			Expect(metric.Host).To(Equal(ip))
			Expect(metric.Source).To(Equal(job))
			Expect(metric.SourceType).To(Equal("cf:containermetric"))
		})

		It("common components", func() {
			event := metric.Event.(SplunkContainerMetric)

			Expect(event.Deployment).To(Equal("cf-warden"))
			Expect(event.JobIndex).To(Equal("85c9ff80-e99b-470b-a194-b397a6e73913"))
			Expect(event.EventType).To(Equal("ContainerMetric"))
			Expect(event.Origin).To(Equal("DEA"))

		})

		It("metric specific data", func() {
			event := metric.Event.(SplunkContainerMetric)

			Expect(event.ApplicationId).To(Equal(applicationId))
			Expect(event.InstanceIndex).To(Equal(instanceIndex))
			Expect(event.CpuPercentage).To(Equal(cpuPercentage))
			Expect(event.MemoryBytes).To(Equal(memoryBytes))
			Expect(event.DiskBytes).To(Equal(diskBytes))
		})
	})
})
