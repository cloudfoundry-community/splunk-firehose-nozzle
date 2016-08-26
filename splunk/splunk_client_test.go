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
	)

	BeforeEach(func() {
		logger = lager.NewLogger("test")
	})

	Context("success response", func() {
		BeforeEach(func() {
			capturedRequest = nil

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

		It("correctly authenticates requests", func() {
			tokenValue := "abc-some-random-token"
			client := NewSplunkClient(tokenValue, testServer.URL, true, logger)
			events := []map[string]interface{}{}
			err := client.Post(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			authValue := capturedRequest.Header.Get("Authorization")
			expectedAuthValue := fmt.Sprintf("Splunk %s", tokenValue)

			Expect(authValue).To(Equal(expectedAuthValue))
		})

		It("sets content type to json", func() {
			client := NewSplunkClient("token", testServer.URL, true, logger)
			events := []map[string]interface{}{}
			err := client.Post(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			contentType := capturedRequest.Header.Get("Content-Type")
			Expect(contentType).To(Equal("application/json"))
		})

		It("posts batch event json", func() {
			client := NewSplunkClient("token", testServer.URL, true, logger)
			event1 := map[string]interface{}{"event": map[string]interface{}{
				"greeting": "hello world",
			}}
			event2 := map[string]interface{}{"event": map[string]interface{}{
				"greeting": "hello mars",
			}}
			event3 := map[string]interface{}{"event": map[string]interface{}{
				"greeting": "hello pluto",
			}}

			events := []map[string]interface{}{event1, event2, event3}
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
			events := []map[string]interface{}{}
			err := client.Post(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest.URL.Path).To(Equal("/services/collector"))
		})
	})

	It("returns error on bad splunk host", func() {
		client := NewSplunkClient("token", ":", true, logger)
		events := []map[string]interface{}{}
		err := client.Post(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("protocol"))
	})

	It("Returns error on non-2xx response", func() {
		testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(500)
			writer.Write([]byte("Internal server error"))
		}))

		client := NewSplunkClient("token", testServer.URL, true, logger)
		events := []map[string]interface{}{}
		err := client.Post(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("500"))
	})

	It("Returns error from http client", func() {
		client := NewSplunkClient("token", "foo://example.com", true, logger)
		events := []map[string]interface{}{}
		err := client.Post(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("foo"))
	})
})
