package nozzle_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/pivotal-golang/lager"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/nozzle"
)

var _ = Describe("Nozzle", func() {
	var (
		nozzle Nozzle
		logger lager.Logger

		eventChannel         chan *events.Envelope
		errorChannel         chan error
		capturedSplunkEvents []*SplunkEvent

		origin        string
		deployment    string
		job           string
		index         string
		ip            string
		timestampNano int64
		envelope      *events.Envelope
		eventType     events.Envelope_EventType
	)

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		eventChannel = make(chan *events.Envelope)
		errorChannel = make(chan error, 1)

		capturedSplunkEvents = nil
		nozzle = NewSplunkForwarder(&MockSplunkClient{
			PostBatchFn: func(events []*SplunkEvent) error {
				capturedSplunkEvents = events
				return nil
			},
		}, eventChannel, errorChannel, logger)

		timestampNano = 1467040874046121775
		deployment = "cf-warden"
		index = "0"
		ip = "10.244.0.22"
		envelope = &events.Envelope{
			Origin:     &origin,
			EventType:  &eventType,
			Timestamp:  &timestampNano,
			Deployment: &deployment,
			Job:        &job,
			Index:      &index,
			Ip:         &ip,
		}
	})

	It("returns error on error channel", func() {
		go func() {
			errorChannel <- errors.New("Fail")
		}()
		err := nozzle.Run(time.Millisecond)

		Expect(err).To(Equal(errors.New("Fail")))
	})

	Context("Envelope HttpStartStop", func() {
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

		It("posts envelope", func() {
			go func() { nozzle.Run(time.Millisecond * 100) }()

			eventChannel <- envelope

			Eventually(func() []*SplunkEvent {
				return capturedSplunkEvents
			}).ShouldNot(BeNil())
			Expect(capturedSplunkEvents).To(HaveLen(1))
			Expect(
				capturedSplunkEvents[0].Event.(SplunkHttpStartStopMetric).EventType,
			).To(Equal("HttpStartStop"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildHttpStartStopMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkHttpStartStopMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("HttpStartStop"))
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
	})

	Context("Envelope LogMessage", func() {
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

		It("posts envelope", func() {
			go func() { nozzle.Run(time.Millisecond) }()

			eventChannel <- envelope

			Eventually(func() []*SplunkEvent {
				return capturedSplunkEvents
			}).ShouldNot(BeNil())
			Expect(capturedSplunkEvents).To(HaveLen(1))
			Expect(
				capturedSplunkEvents[0].Event.(SplunkLogMessageMetric).EventType,
			).To(Equal("LogMessage"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildLogMessageMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkLogMessageMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("LogMessage"))
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
	})

	Context("Envelope ValueMetric", func() {
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

		It("posts envelope", func() {
			go func() { nozzle.Run(time.Millisecond) }()

			eventChannel <- envelope

			Eventually(func() []*SplunkEvent {
				return capturedSplunkEvents
			}).ShouldNot(BeNil())
			Expect(capturedSplunkEvents).To(HaveLen(1))
			Expect(
				capturedSplunkEvents[0].Event.(SplunkValueMetric).EventType,
			).To(Equal("ValueMetric"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildValueMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkValueMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("ValueMetric"))
			})

			It("metric specific data", func() {
				event := metric.Event.(SplunkValueMetric)

				Expect(event.Name).To(Equal(name))
				Expect(event.Value).To(Equal(value))
				Expect(event.Unit).To(Equal(unit))
			})
		})
	})

	Context("Envelope CounterEvent", func() {
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

		It("posts envelope", func() {
			go func() { nozzle.Run(time.Millisecond) }()

			eventChannel <- envelope

			Eventually(func() []*SplunkEvent {
				return capturedSplunkEvents
			}).ShouldNot(BeNil())
			Expect(capturedSplunkEvents).To(HaveLen(1))
			Expect(
				capturedSplunkEvents[0].Event.(SplunkCounterEventMetric).EventType,
			).To(Equal("CounterEvent"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildCounterEventMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkCounterEventMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("CounterEvent"))
			})

			It("metric specific data", func() {
				event := metric.Event.(SplunkCounterEventMetric)

				Expect(event.Name).To(Equal(name))
				Expect(event.Delta).To(Equal(delta))
				Expect(event.Total).To(Equal(total))
			})
		})
	})

	Context("Envelope Error", func() {
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
		})

		It("posts envelope", func() {
			go func() { nozzle.Run(time.Millisecond) }()

			eventChannel <- envelope

			Eventually(func() []*SplunkEvent {
				return capturedSplunkEvents
			}).ShouldNot(BeNil())
			Expect(
				capturedSplunkEvents[0].Event.(SplunkErrorMetric).EventType,
			).To(Equal("Error"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildErrorMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkErrorMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("Error"))
			})

			It("metric specific data", func() {
				event := metric.Event.(SplunkErrorMetric)

				Expect(event.Source).To(Equal(source))
				Expect(event.Code).To(Equal(code))
				Expect(event.Message).To(Equal(message))
			})
		})
	})

	Context("Envelope ContainerMetric", func() {
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

		It("posts envelope", func() {
			go func() { nozzle.Run(time.Millisecond) }()

			eventChannel <- envelope

			Eventually(func() []*SplunkEvent {
				return capturedSplunkEvents
			}).ShouldNot(BeNil())
			Expect(capturedSplunkEvents).To(HaveLen(1))
			Expect(
				capturedSplunkEvents[0].Event.(SplunkContainerMetric).EventType,
			).To(Equal("ContainerMetric"))
		})

		Context("translates", func() {
			var metric *SplunkEvent
			BeforeEach(func() {
				metric = BuildContainerMetric(envelope)
			})

			It("metadata", func() {
				eventTimeSeconds := "1467040874.046"
				Expect(metric.Time).To(Equal(eventTimeSeconds))
				Expect(metric.Host).To(Equal(ip))
				Expect(metric.Source).To(Equal(job))
			})

			It("common components", func() {
				event := metric.Event.(SplunkContainerMetric)

				Expect(event.Deployment).To(Equal("cf-warden"))
				Expect(event.Index).To(Equal("0"))
				Expect(event.EventType).To(Equal("ContainerMetric"))
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
})

type MockSplunkClient struct {
	PostSingleFn func(event *SplunkEvent) error
	PostBatchFn  func(event []*SplunkEvent) error
}

func (m *MockSplunkClient) PostSingle(event *SplunkEvent) error {
	if m.PostSingleFn != nil {
		return m.PostSingleFn(event)
	}
	return nil
}

func (m *MockSplunkClient) PostBatch(event []*SplunkEvent) error {
	if m.PostBatchFn != nil {
		return m.PostBatchFn(event)
	}
	return nil
}
