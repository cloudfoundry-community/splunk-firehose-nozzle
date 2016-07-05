package writernozzle_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/writernozzle"
)

var _ = Describe("WriterClient", func() {
	var (
		serializer *WriterEventSerializer
	)

	BeforeEach(func() {
		serializer = NewWriterEventSerializer()
	})

	It("serializes events to []byte", func() {
		origin := "gorouter"
		eventType := events.Envelope_Error
		timestamp := int64(1467040874046121775)
		deployment := "cf-warden"
		job := "runner_z1"
		index := "0"
		ip := "10.244.0.22"

		envelope := &events.Envelope{
			Origin:     &origin,
			EventType:  &eventType,
			Timestamp:  &timestamp,
			Deployment: &deployment,
			Job:        &job,
			Index:      &index,
			Ip:         &ip,
		}

		source := "some_source"
		message := "something failed"
		code := int32(42)
		envelopeError := &events.Error{
			Source:  &source,
			Code:    &code,
			Message: &message,
		}

		envelope.Error = envelopeError

		serialized := serializer.BuildErrorEvent(envelope)

		Expect(fmt.Sprintf("%+v\n", envelope)).To(Equal(string(serialized.([]byte))))
	})
})
