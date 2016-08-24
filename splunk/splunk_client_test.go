package splunk_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/splunk"
)

var _ = Describe("SplunkClient", func() {
	var (
		testServer      *httptest.Server
		capturedRequest *http.Request
		capturedBody    []byte
		splunkResponse  []byte
		logger          lager.Logger

		emptyJson []byte
	)

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		emptyJson = []byte("{}")
	})

	Context("success response", func() {
		BeforeEach(func() {
			splunkResponse = []byte("{}")
			testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				capturedRequest = request
				body, err := ioutil.ReadAll(request.Body)
				if err != nil {
					panic(err)
				}
				capturedBody = body

				writer.Write(splunkResponse)
			}))
		})

		AfterEach(func() {
			testServer.Close()
		})

		FIt("correctly authenticates requests", func() {
			tokenValue := "abc-some-random-token"
			client := NewSplunkClient(tokenValue, testServer.URL, true, logger)
			events := []*[]byte{&emptyJson}
			err := client.Post(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			authValue := capturedRequest.Header.Get("Authorization")
			expectedAuthValue := fmt.Sprintf("Splunk %s", tokenValue)

			Expect(authValue).To(Equal(expectedAuthValue))
		})

		FIt("sets content type to json", func() {
			client := NewSplunkClient("token", testServer.URL, true, logger)
			events := []*[]byte{&emptyJson}
			err := client.Post(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			contentType := capturedRequest.Header.Get("Content-Type")
			Expect(contentType).To(Equal("application/json"))
		})

		FIt("posts batch event json", func() {
			client := NewSplunkClient("token", testServer.URL, true, logger)
			event1 := []byte(`{"event":{"greeting":"hello world"}}`)
			event2 := []byte(`{"event":{"greeting":"hello mars"}}`)
			event3 := []byte(`{"event":{"greeting":"hello pluto"}}`)

			events := []*[]byte{&event1, &event2, &event3}
			err := client.Post(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			expectedPayload := strings.TrimSpace(`
{"event":{"greeting":"hello world"}}

{"event":{"greeting":"hello mars"}}

{"event":{"greeting":"hello pluto"}}
`)
			Expect(string(capturedBody)).To(Equal(expectedPayload))
		})

		It("posts to correct endpoint", func() {
			client := NewSplunkClient("token", testServer.URL, true, logger)
			events := []interface{}{&SplunkEvent{}}
			err := client.PostBatch(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest.URL.Path).To(Equal("/services/collector"))
		})
	})

	It("returns error on bad splunk host", func() {
		client := NewSplunkClient("token", ":", true, logger)

		events := []interface{}{&SplunkEvent{}}
		err := client.PostBatch(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("protocol"))
	})

	It("Returns error on non-2xx response", func() {
		testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(500)
			writer.Write([]byte("Internal server error"))
		}))

		client := NewSplunkClient("token", testServer.URL, true, logger)
		events := []interface{}{&SplunkEvent{}}
		err := client.PostBatch(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("500"))
	})

	It("Returns error from http client", func() {
		client := NewSplunkClient("token", "foo://example.com", true, logger)
		events := []interface{}{&SplunkEvent{}}
		err := client.PostBatch(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("foo"))
	})
})
