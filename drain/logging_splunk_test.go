package drain_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"strconv"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry/sonde-go/events"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/drain"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/testing"
)

var _ = Describe("LoggingSplunk", func() {

	var (
		origin        string
		deployment    string
		job           string
		jobIndex      string
		ip            string
		timestampNano int64
		envelope      *events.Envelope
		eventType     events.Envelope_EventType

		loggingMemory *drain.LoggingMemory
		logging       *drain.LoggingSplunk

		event      map[string]interface{}
		logger     lager.Logger
		mockClient *testing.MockSplunkClient
		routing    *eventRouting.EventRouting
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

		//using routing to serialize envelope
		loggingMemory = drain.NewLoggingMemory()
		routing = eventRouting.NewEventRouting(caching.NewCachingEmpty(), loggingMemory)
		routing.SetupEventRouting("HttpStartStop,LogMessage,ValueMetric,CounterEvent,ContainerMetric,Error")

		mockClient = &testing.MockSplunkClient{}

		logger = lager.NewLogger("test")
		logging = drain.NewLoggingSplunk(logger, mockClient, time.Millisecond)
	})

	It("sends events to client", func() {
		eventType = events.Envelope_Error
		routing.RouteEvent(envelope)

		logging.Connect()
		logging.ShipEvents(loggingMemory.Events[0], loggingMemory.Messages[0])

		Eventually(func() []map[string]interface{} {
			return mockClient.CapturedEvents
		}).Should(HaveLen(1))
	})

	It("job_index is present, index is not", func() {
		eventType = events.Envelope_Error
		routing.RouteEvent(envelope)

		logging.Connect()
		logging.ShipEvents(loggingMemory.Events[0], loggingMemory.Messages[0])

		Eventually(func() []map[string]interface{} {
			return mockClient.CapturedEvents
		}).Should(HaveLen(1))

		event = mockClient.CapturedEvents[0]

		data := event["event"].(map[string]interface{})
		Expect(data).NotTo(HaveKey("index"))

		index := data["job_index"]
		Expect(index).To(Equal(jobIndex))
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

		})

		BeforeEach(func() {
			routing.RouteEvent(envelope)

			logging.Connect()
			logging.ShipEvents(loggingMemory.Events[0], loggingMemory.Messages[0])

			Eventually(func() []map[string]interface{} {
				return mockClient.CapturedEvents
			}).Should(HaveLen(1))

			event = mockClient.CapturedEvents[0]
		})

		It("metadata", func() {
			Expect(event["host"]).To(Equal(ip))
			Expect(event["source"]).To(Equal(job))
			Expect(event["sourcetype"]).To(Equal("cf:httpstartstop"))
		})

		It("uses current time without event timestamp", func() {
			eventTime, err := strconv.Atoi(event["time"].(string))

			Expect(err).To(BeNil())
			Expect(time.Now().Unix()).To(BeNumerically("~", eventTime, 2))
		})

		//todo: manual test around start/stop time: number of digits?
		It("adds fields to payload.event", func() {
			eventContents := event["event"].(map[string]interface{})

			Expect(eventContents["method"]).To(Equal("GET"))
			Expect(eventContents["remote_addr"]).To(Equal(remoteAddress))
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
		})

		BeforeEach(func() {
			routing.RouteEvent(envelope)

			logging.Connect()
			logging.ShipEvents(loggingMemory.Events[0], loggingMemory.Messages[0])

			Eventually(func() []map[string]interface{} {
				return mockClient.CapturedEvents
			}).Should(HaveLen(1))

			event = mockClient.CapturedEvents[0]
		})

		It("metadata", func() {
			Expect(event["host"]).To(Equal(ip))
			Expect(event["source"]).To(Equal(job))
			Expect(event["sourcetype"]).To(Equal("cf:logmessage"))
		})

		It("uses event timestamp", func() {
			eventTimeSeconds := "1467128185.055"
			Expect(event["time"]).To(Equal(eventTimeSeconds))
		})

		It("adds fields to payload.event", func() {
			eventContents := event["event"].(map[string]interface{})

			Expect(eventContents["cf_app_id"]).To(Equal(appId))
			Expect(eventContents["message_type"]).To(Equal("OUT"))
		})

		It("adds message", func() {
			eventContents := event["event"].(map[string]interface{})

			Expect(eventContents["msg"]).To(Equal("App debug log message"))
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
		})

		BeforeEach(func() {
			routing.RouteEvent(envelope)

			logging.Connect()
			logging.ShipEvents(loggingMemory.Events[0], loggingMemory.Messages[0])

			Eventually(func() []map[string]interface{} {
				return mockClient.CapturedEvents
			}).Should(HaveLen(1))

			event = mockClient.CapturedEvents[0]
		})

		It("metadata", func() {
			Expect(event["host"]).To(Equal(ip))
			Expect(event["source"]).To(Equal(job))
			Expect(event["sourcetype"]).To(Equal("cf:valuemetric"))
		})

		It("uses current time without event timestamp", func() {
			eventTime, err := strconv.Atoi(event["time"].(string))

			Expect(err).To(BeNil())
			Expect(time.Now().Unix()).To(BeNumerically("~", eventTime, 2))
		})

		It("adds fields to payload.event", func() {
			eventContents := event["event"].(map[string]interface{})

			Expect(eventContents["unit"]).To(Equal("ms"))
			Expect(eventContents["value"]).To(Equal(1581.0))
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
		})

		BeforeEach(func() {
			routing.RouteEvent(envelope)

			logging.Connect()
			logging.ShipEvents(loggingMemory.Events[0], loggingMemory.Messages[0])

			Eventually(func() []map[string]interface{} {
				return mockClient.CapturedEvents
			}).Should(HaveLen(1))

			event = mockClient.CapturedEvents[0]
		})

		It("metadata", func() {
			Expect(event["host"]).To(Equal(ip))
			Expect(event["source"]).To(Equal(job))
			Expect(event["sourcetype"]).To(Equal("cf:counterevent"))
		})

		It("uses current time without event timestamp", func() {
			eventTime, err := strconv.Atoi(event["time"].(string))

			Expect(err).To(BeNil())
			Expect(time.Now().Unix()).To(BeNumerically("~", eventTime, 2))
		})

		It("adds fields to payload.event", func() {
			eventContents := event["event"].(map[string]interface{})

			Expect(eventContents["origin"]).To(Equal("gorouter"))
			Expect(eventContents["name"]).To(Equal("registry_message.uaa"))
		})
	})

	Context("envelope Error", func() {
		var source, message string
		var code int32
		var envelopeError *events.Error

		BeforeEach(func() {
			source = "some_source"
			message = "Something failed"
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
		})

		BeforeEach(func() {
			routing.RouteEvent(envelope)

			logging.Connect()
			logging.ShipEvents(loggingMemory.Events[0], loggingMemory.Messages[0])

			Eventually(func() []map[string]interface{} {
				return mockClient.CapturedEvents
			}).Should(HaveLen(1))

			event = mockClient.CapturedEvents[0]
		})

		It("metadata", func() {
			Expect(event["host"]).To(Equal(ip))
			Expect(event["source"]).To(Equal(job))
			Expect(event["sourcetype"]).To(Equal("cf:error"))
		})

		It("uses current time without event timestamp", func() {
			eventTime, err := strconv.Atoi(event["time"].(string))

			Expect(err).To(BeNil())
			Expect(time.Now().Unix()).To(BeNumerically("~", eventTime, 2))
		})

		It("adds fields to payload.event", func() {
			eventContents := event["event"].(map[string]interface{})

			Expect(eventContents["code"]).To(BeNumerically("==", 42))
		})

		It("adds message", func() {
			eventContents := event["event"].(map[string]interface{})

			Expect(eventContents["msg"]).To(Equal("Something failed"))
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
		})

		BeforeEach(func() {
			routing.RouteEvent(envelope)

			logging.Connect()
			logging.ShipEvents(loggingMemory.Events[0], loggingMemory.Messages[0])

			Eventually(func() []map[string]interface{} {
				return mockClient.CapturedEvents
			}).Should(HaveLen(1))

			event = mockClient.CapturedEvents[0]
		})

		It("metadata", func() {
			Expect(event["host"]).To(Equal(ip))
			Expect(event["source"]).To(Equal(job))
			Expect(event["sourcetype"]).To(Equal("cf:containermetric"))
		})

		It("uses current time without event timestamp", func() {
			eventTime, err := strconv.Atoi(event["time"].(string))

			Expect(err).To(BeNil())
			Expect(time.Now().Unix()).To(BeNumerically("~", eventTime, 2))
		})

		It("adds fields to payload.event", func() {
			eventContents := event["event"].(map[string]interface{})

			Expect(eventContents["cpu_percentage"]).To(Equal(cpuPercentage))
			Expect(eventContents["memory_bytes"]).To(Equal(memoryBytes))
		})
	})
})
