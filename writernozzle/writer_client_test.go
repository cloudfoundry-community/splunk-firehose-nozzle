package writernozzle_test

import (
	"bytes"
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/writernozzle"
)

var _ = Describe("WriterClient", func() {
	var (
		writerClient *WriterClient
		buffer       *bytes.Buffer
	)

	BeforeEach(func() {
		buffer = &bytes.Buffer{}
		writerClient = NewWriterClient(buffer)
	})

	It("writes events to stream", func() {
		events := []interface{}{
			[]byte("first"),
			[]byte("second"),
		}

		writerClient.PostBatch(events)

		Expect(buffer.String()).To(Equal("firstsecond"))
	})

	It("returns error if problem writing bytes", func() {
		writerClient = NewWriterClient(&ErrorWriter{})

		err := writerClient.PostBatch([]interface{}{
			[]byte(""),
		})

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("some error"))
	})
})

type ErrorWriter struct{}

func (e *ErrorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("some error")
}
