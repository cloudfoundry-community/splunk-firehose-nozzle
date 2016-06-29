package nozzle_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/pivotal-golang/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/nozzle"
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

		Context("post single", func() {
			It("posts event json", func() {
				client := NewSplunkClient("token", testServer.URL, true, logger)
				err := client.PostSingle(&SplunkEvent{
					Event: map[string]string{
						"message": "hello world",
					},
				})

				Expect(err).To(BeNil())
				Expect(capturedRequest).NotTo(BeNil())
				Expect(string(capturedBody)).To(Equal(`{"event":{"message":"hello world"}}`))
			})
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("correctly authenticates requests", func() {
			tokenValue := "abc-some-random-token"
			client := NewSplunkClient(tokenValue, testServer.URL, true, logger)
			events := []*SplunkEvent{&SplunkEvent{}}
			err := client.PostBatch(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			authValue := capturedRequest.Header.Get("Authorization")
			expectedAuthValue := fmt.Sprintf("Splunk %s", tokenValue)

			Expect(authValue).To(Equal(expectedAuthValue))
		})

		It("sets content type to json", func() {
			client := NewSplunkClient("token", testServer.URL, true, logger)
			events := []*SplunkEvent{&SplunkEvent{}}
			err := client.PostBatch(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			contentType := capturedRequest.Header.Get("Content-Type")
			Expect(contentType).To(Equal("application/json"))
		})

		It("posts batch event json", func() {
			client := NewSplunkClient("token", testServer.URL, true, logger)
			events := []*SplunkEvent{
				&SplunkEvent{
					Event: map[string]string{"greeting": "hello world"},
				},
				&SplunkEvent{
					Event: map[string]string{"message": "hello mars"},
				},
				&SplunkEvent{
					Event: map[string]string{"message": "hello pluto"},
				},
			}
			err := client.PostBatch(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			expectedPayload := strings.TrimSpace(`
{"event":{"greeting":"hello world"}}

{"event":{"message":"hello mars"}}

{"event":{"message":"hello pluto"}}
`)
			Expect(string(capturedBody)).To(Equal(expectedPayload))
		})

		It("posts to correct endpoint", func() {
			client := NewSplunkClient("token", testServer.URL, true, logger)
			events := []*SplunkEvent{&SplunkEvent{}}
			err := client.PostBatch(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest.URL.Path).To(Equal("/services/collector"))
		})
	})

	It("returns error on bad splunk host", func() {
		client := NewSplunkClient("token", ":", true, logger)

		events := []*SplunkEvent{&SplunkEvent{}}
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
		events := []*SplunkEvent{&SplunkEvent{}}
		err := client.PostBatch(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("500"))
	})

	It("Returns error from http client", func() {
		client := NewSplunkClient("token", "foo://example.com", true, logger)
		events := []*SplunkEvent{&SplunkEvent{}}
		err := client.PostBatch(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("foo"))
	})
})
