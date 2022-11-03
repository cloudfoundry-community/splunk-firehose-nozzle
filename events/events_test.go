package events_test

import (
	"math"

	fevents "github.com/cloudfoundry-community/splunk-firehose-nozzle/events"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/testing"
	. "github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events", func() {
	var (
		fcache *testing.MemoryCacheMock
		event  *fevents.Event
		msg    *Envelope
	)

	BeforeEach(func() {
		fcache = &testing.MemoryCacheMock{}
		msg = NewLogMessage()
		event = fevents.LogMessage(msg)
		event.AnnotateWithEnvelopeData(msg, &fevents.Config{})
	})

	It("HttpStart", func() {
		msg = NewHttpStart()
		evt := fevents.HttpStart(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["timestamp"]).To(Equal(timestamp))
		Expect(evt.Fields["request_id"]).To(Equal(uuidStr))
		Expect(evt.Fields["method"]).To(Equal(methodStr))
		Expect(evt.Fields["uri"]).To(Equal(uri))
		Expect(evt.Fields["remote_addr"]).To(Equal(remoteAddr))
		Expect(evt.Fields["user_agent"]).To(Equal(userAgent))
		Expect(evt.Fields["parent_request_id"]).To(Equal(uuidStr))
		Expect(evt.Fields["cf_app_id"]).To(Equal(uuidStr))
		Expect(evt.Fields["instance_index"]).To(Equal(instanceIdx))
		Expect(evt.Fields["instance_id"]).To(Equal(instanceId))
	})

	It("HttpStop", func() {
		msg = NewHttpStop()
		evt := fevents.HttpStop(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["timestamp"]).To(Equal(timestamp))
		Expect(evt.Fields["uri"]).To(Equal(uri))
		Expect(evt.Fields["request_id"]).To(Equal(uuidStr))
		Expect(evt.Fields["peer_type"]).To(Equal(peerTypeStr))
		Expect(evt.Fields["status_code"]).To(Equal(statusCode))
		Expect(evt.Fields["content_length"]).To(Equal(contentLength))
		Expect(evt.Fields["cf_app_id"]).To(Equal(uuidStr))
	})

	It("HttpStartStop", func() {
		msg = NewHttpStartStop()
		evt := fevents.HttpStartStop(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["cf_app_id"]).To(Equal(uuidStr))
		Expect(evt.Fields["content_length"]).To(Equal(contentLength))
		Expect(evt.Fields["instance_index"]).To(Equal(instanceIdx))
		Expect(evt.Fields["instance_id"]).To(Equal(instanceId))
		Expect(evt.Fields["method"]).To(Equal(methodStr))
		Expect(evt.Fields["peer_type"]).To(Equal(peerTypeStr))
		Expect(evt.Fields["remote_addr"]).To(Equal(remoteAddr))
		Expect(evt.Fields["request_id"]).To(Equal(uuidStr))
		Expect(evt.Fields["start_timestamp"]).To(Equal(timestamp))
		Expect(evt.Fields["stop_timestamp"]).To(Equal(timestamp))
		Expect(evt.Fields["status_code"]).To(Equal(statusCode))
		Expect(evt.Fields["uri"]).To(Equal(uri))
		Expect(evt.Fields["user_agent"]).To(Equal(userAgent))
		Expect(evt.Fields["forwarded"]).To(BeNil())
	})

	It("ValueMetric", func() {
		msg = NewValueMetric()
		evt := fevents.ValueMetric(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["name"]).To(Equal(name))
		Expect(evt.Fields["value"]).To(Equal(value))
		Expect(evt.Fields["unit"]).To(Equal(unit))
	})

	It("ValueMetric NaN", func() {
		msg = NewValueMetric()
		nan := math.NaN()
		msg.ValueMetric.Value = &nan
		evt := fevents.ValueMetric(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["name"]).To(Equal(name))
		Expect(evt.Fields["value"]).To(Equal("NaN"))
		Expect(evt.Fields["unit"]).To(Equal(unit))
	})

	It("ValueMetric +Infinity", func() {
		msg = NewValueMetric()
		inf := math.Inf(1)
		msg.ValueMetric.Value = &inf
		evt := fevents.ValueMetric(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["name"]).To(Equal(name))
		Expect(evt.Fields["value"]).To(Equal("Infinity"))
		Expect(evt.Fields["unit"]).To(Equal(unit))
	})

	It("ValueMetric -Infinity", func() {
		msg = NewValueMetric()
		inf := math.Inf(-1)
		msg.ValueMetric.Value = &inf
		evt := fevents.ValueMetric(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["name"]).To(Equal(name))
		Expect(evt.Fields["value"]).To(Equal("-Infinity"))
		Expect(evt.Fields["unit"]).To(Equal(unit))
	})

	It("CounterEvent", func() {
		msg = NewCounterEvent()
		evt := fevents.CounterEvent(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["name"]).To(Equal(name))
		Expect(evt.Fields["delta"]).To(Equal(delta))
		Expect(evt.Fields["total"]).To(Equal(total))
	})

	It("ErrorEvent", func() {
		msg = NewErrorEvent()
		evt := fevents.ErrorEvent(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(message))
		Expect(evt.Fields["code"]).To(Equal(code))
		Expect(evt.Fields["source"]).To(Equal(source))
	})

	It("ContainerMetric", func() {
		msg = NewContainerMetric()
		evt := fevents.ContainerMetric(msg)
		Expect(evt).ToNot(BeNil())
		Expect(evt.Fields).ToNot(BeNil())
		Expect(evt.Msg).To(Equal(""))
		Expect(evt.Fields["cf_app_id"]).To(Equal(uuidStr))
		Expect(evt.Fields["cpu_percentage"]).To(Equal(cpuPercentage))
		Expect(evt.Fields["disk_bytes"]).To(Equal(diskBytes))
		Expect(evt.Fields["disk_bytes_quota"]).To(Equal(diskBytesQuota))
		Expect(evt.Fields["memory_bytes"]).To(Equal(memoryBytes))
		Expect(evt.Fields["memory_bytes_quota"]).To(Equal(memoryBytesQuota))
		Expect(evt.Fields["instance_index"]).To(Equal(instanceIdx))
	})

	Context("given a envelope", func() {
		It("should give us what we want", func() {
			Expect(event.Fields["origin"]).To(Equal("yomomma__0"))
			Expect(event.Fields["cf_app_id"]).To(Equal("eea38ba5-53a5-4173-9617-b442d35ec2fd"))
			Expect(event.Fields["timestamp"]).To(Equal(int64(1)))
			Expect(event.Fields["source_type"]).To(Equal("Kehe"))
			Expect(event.Fields["message_type"]).To(Equal("OUT"))
			Expect(event.Fields["source_instance"]).To(Equal(">9000"))
			Expect(event.Msg).To(Equal("Help, I'm a rock! Help, I'm a rock! Help, I'm a cop! Help, I'm a cop!"))
		})
	})

	Context("given metadata", func() {
		It("Should give us the right metadata", func() {
			event.AnnotateWithCFMetaData()
			Expect(event.Fields["event_type"]).To(Equal(event.Type))

		})

	})

	Context("given Application Metadata", func() {
		It("Should give us the right Application metadata", func() {
			fcache.SetIgnoreApp(true)
			var config = &fevents.Config{
				AddAppName:   true,
				AddOrgName:   true,
				AddOrgGuid:   true,
				AddSpaceName: true,
				AddSpaceGuid: true,
				AddTags:      true,
			}
			event.AnnotateWithAppData(fcache, config)
			Expect(event.Fields["cf_app_name"]).To(Equal("testing-app"))
			Expect(event.Fields["cf_space_id"]).To(Equal("f964a41c-76ac-42c1-b2ba-663da3ec22d6"))
			Expect(event.Fields["cf_space_name"]).To(Equal("testing-space"))
			Expect(event.Fields["cf_org_id"]).To(Equal("f964a41c-76ac-42c1-b2ba-663da3ec22d7"))
			Expect(event.Fields["cf_org_name"]).To(Equal("testing-org"))

			event.AnnotateWithEnvelopeData(msg, config)
			Expect(event.Fields["tags"]).To(Equal(msg.GetTags()))
		})
	})

	It("HttpStart", func() {
		var config = &fevents.Config{
			AddAppName:   true,
			AddOrgName:   true,
			AddOrgGuid:   true,
			AddSpaceName: true,
			AddSpaceGuid: true,
		}
		event.AnnotateWithAppData(fcache, config)
		Expect(event.Fields["cf_app_name"]).To(Equal("testing-app"))
		Expect(event.Fields["cf_space_id"]).To(Equal("f964a41c-76ac-42c1-b2ba-663da3ec22d6"))
		Expect(event.Fields["cf_space_name"]).To(Equal("testing-space"))
		Expect(event.Fields["cf_org_id"]).To(Equal("f964a41c-76ac-42c1-b2ba-663da3ec22d7"))
		Expect(event.Fields["cf_org_name"]).To(Equal("testing-org"))
	})

	Context("ParseSelectedEvents, empty select events passed in", func() {
		It("should return a hash of only the default event", func() {
			results, err := fevents.ParseSelectedEvents("")
			Ω(err).ShouldNot(HaveOccurred())
			expected := map[string]bool{"LogMessage": true}
			Expect(results).To(Equal(expected))
		})
	})

	Context("ParseSelectedEvents, bogus event names", func() {
		It("should err out", func() {
			_, err := fevents.ParseSelectedEvents("bogus, invalid")
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("ParseSelectedEvents, valid event names", func() {
		It("should return a hash of events", func() {
			expected := map[string]bool{
				"HttpStartStop": true,
				"CounterEvent":  true,
			}
			results, err := fevents.ParseSelectedEvents("HttpStartStop,CounterEvent")
			Ω(err).ShouldNot(HaveOccurred())
			Expect(results).To(Equal(expected))
		})
	})

	Context("ParseSelectedEvents, valid event names in string of list", func() {
		It("should return a hash of events", func() {
			expected := map[string]bool{
				"HttpStartStop": true,
				"CounterEvent":  true,
			}
			results, err := fevents.ParseSelectedEvents(`["HttpStartStop","CounterEvent"]`)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(results).To(Equal(expected))
		})
	})

	Context("AuthorizedEvents", func() {
		It("should return right list of authorized events", func() {
			Expect(fevents.AuthorizedEvents()).To(Equal("ContainerMetric, CounterEvent, Error, HttpStart, HttpStartStop, HttpStop, LogMessage, ValueMetric"))
		})
	})

	Describe("ParseExtraFields", func() {
		Context("called with a empty string", func() {
			It("should return a empty hash", func() {
				expected := map[string]string{}
				Expect(fevents.ParseExtraFields("")).To(Equal(expected))
			})
		})

		Context("called with extra events", func() {
			It("should return a hash with the events we want", func() {
				expected := map[string]string{"env": "dev", "kehe": "wakawaka"}
				extraEvents := "env:dev,kehe:wakawaka"
				Expect(fevents.ParseExtraFields(extraEvents)).To(Equal(expected))
			})
		})

		Context("called with extra events with weird whitespace", func() {
			It("should return a hash with the events we want", func() {
				expected := map[string]string{"env": "dev", "kehe": "wakawaka"}
				extraEvents := "    env:      \ndev,      kehe:wakawaka   "
				Expect(fevents.ParseExtraFields(extraEvents)).To(Equal(expected))
			})
		})

		Context("called with extra events with to many values to a kv pair", func() {
			It("should return a error", func() {
				extraEvents := "to:many:values"
				_, err := fevents.ParseExtraFields(extraEvents)
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
