package sink_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/lager"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/testing"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/sink"
)

var _ = Describe("LoggingSplunk", func() {
	var (
		capturedEvents []map[string]interface{}
		mockClient     *testing.MockSplunkClient

		sink *SplunkSink
	)

	BeforeEach(func() {
		capturedEvents = nil
		mockClient = &testing.MockSplunkClient{}

		sink = NewSplunkSink("splunk-nozzle_z1", "0", "127.0.0.1", mockClient)
	})

	It("posts to splunk", func() {
		message := lager.LogFormat{}

		Expect(mockClient.CapturedEvents).To(BeNil())

		sink.Log(message)
		sink.Log(message)

		Expect(mockClient.CapturedEvents).To(HaveLen(2))
	})

	It("translates log message metadata to splunk format", func() {
		message := lager.LogFormat{
			Timestamp: "1473180363",
			Source:    "splunk-nozzle-logger",
			Message:   "Failure",
			LogLevel:  lager.ERROR,
		}

		sink.Log(message)

		Expect(mockClient.CapturedEvents).To(HaveLen(1))
		envelope := mockClient.CapturedEvents[0]

		Expect(envelope["time"]).To(Equal("1473180363"))

		event := envelope["event"].(map[string]interface{})
		Expect(event["logger_source"]).To(Equal("splunk-nozzle-logger"))
		Expect(event["log_level"]).To(Equal(2))
	})

	It("translates log message payload to splunk format", func() {
		message := lager.LogFormat{
			Timestamp: "1473180363",
			Source:    "splunk-nozzle-logger",
			Message:   "Failure",
			LogLevel:  lager.ERROR,
			Data: lager.Data{
				"foo": "bar",
				"baz": 42,
			},
		}

		sink.Log(message)

		Expect(mockClient.CapturedEvents).To(HaveLen(1))
		envelope := mockClient.CapturedEvents[0]

		Expect(envelope["time"]).To(Equal("1473180363"))

		event := envelope["event"].(map[string]interface{})
		Expect(event["message"]).To(Equal("Failure"))

		data := event["data"].(map[string]interface{})
		Expect(data["foo"]).To(Equal("bar"))
		Expect(data["baz"]).To(Equal(42))
	})

	It("adds expected Splunk fields", func() {
		message := lager.LogFormat{}

		sink.Log(message)

		Expect(mockClient.CapturedEvents).To(HaveLen(1))
		envelope := mockClient.CapturedEvents[0]

		Expect(envelope["sourcetype"]).To(Equal("cf:splunknozzle"))
		Expect(envelope["host"]).To(Equal("127.0.0.1"))
		Expect(envelope["source"]).To(Equal("splunk-nozzle_z1"))

		event := envelope["event"].(map[string]interface{})
		Expect(event["index"]).To(BeNil())
		Expect(event["job_index"]).To(Equal("0"))
		Expect(event["job"]).To(Equal("splunk-nozzle_z1"))
		Expect(event["ip"]).To(Equal("127.0.0.1"))
		Expect(event["origin"]).To(Equal("splunk_nozzle"))
	})
})
