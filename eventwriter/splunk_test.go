package eventwriter_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"code.cloudfoundry.org/lager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/eventwriter"
)

var _ = Describe("Splunk", func() {
	var (
		testServer      *httptest.Server
		capturedRequest *http.Request
		capturedBody    []byte
		splunkResponse  []byte
		logger          lager.Logger
		config          *SplunkConfig
	)

	BeforeEach(func() {
		logger = lager.NewLogger("test")
		config = &SplunkConfig{
			Token:    "token",
			Index:    "",
			Fields:   nil,
			SkipSSL:  true,
			Endpoint: "/services/collector",
			Logger:   logger,
		}
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

			config.Host = testServer.URL
		})

		AfterEach(func() {
			testServer.Close()
		})

		It("correctly authenticates requests", func() {
			tokenValue := "abc-some-random-token"
			config.Token = tokenValue

			client := NewSplunk(config)
			events := []map[string]interface{}{}
			err := client.Write(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			authValue := capturedRequest.Header.Get("Authorization")
			expectedAuthValue := fmt.Sprintf("Splunk %s", tokenValue)

			Expect(authValue).To(Equal(expectedAuthValue))
		})

		It("sets content type to json", func() {
			client := NewSplunk(config)
			events := []map[string]interface{}{}
			err := client.Write(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			contentType := capturedRequest.Header.Get("Content-Type")
			Expect(contentType).To(Equal("application/json"))
		})

		It("Writes batch event json", func() {
			client := NewSplunk(config)
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
			err := client.Write(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			expectedPayload := strings.TrimSpace(`
{"event":{"greeting":"hello world"}}

{"event":{"greeting":"hello mars"}}

{"event":{"greeting":"hello pluto"}}
`)
			Expect(string(capturedBody)).To(Equal(expectedPayload))
		})

		It("sets index in splunk payload", func() {
			config.Index = "index_cf"
			client := NewSplunk(config)
			event1 := map[string]interface{}{"event": map[string]interface{}{
				"greeting": "hello world",
			}}
			event2 := map[string]interface{}{"event": map[string]interface{}{
				"greeting": "hello mars",
			}}

			events := []map[string]interface{}{event1, event2}
			err := client.Write(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			expectedPayload := strings.TrimSpace(`
{"event":{"greeting":"hello world"},"index":"index_cf"}

{"event":{"greeting":"hello mars"},"index":"index_cf"}
`)
			Expect(string(capturedBody)).To(Equal(expectedPayload))
		})

		It("adds fields to splunk palylaod", func() {
			fields := map[string]string{
				"foo":   "bar",
				"hello": "world",
			}
			config.Fields = fields

			client := NewSplunk(config)
			event1 := map[string]interface{}{"event": map[string]interface{}{
				"greeting": "hello world",
			}}
			event2 := map[string]interface{}{"event": map[string]interface{}{
				"greeting": "hello mars",
			}}

			events := []map[string]interface{}{event1, event2}
			err := client.Write(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest).NotTo(BeNil())

			expectedPayload := strings.TrimSpace(`
{"event":{"greeting":"hello world"},"fields":{"foo":"bar","hello":"world"}}

{"event":{"greeting":"hello mars"},"fields":{"foo":"bar","hello":"world"}}
`)
			Expect(string(capturedBody)).To(Equal(expectedPayload))

		})

		It("Writes to correct endpoint", func() {
			client := NewSplunk(config)
			events := []map[string]interface{}{}
			err := client.Write(events)

			Expect(err).To(BeNil())
			Expect(capturedRequest.URL.Path).To(Equal("/services/collector"))
		})

		It("Writes to custom endpoint", func() {
			custom_hec_endpoint := "/my/custom/endpoint"
			config.Endpoint = custom_hec_endpoint
			client := NewSplunk(config)
			events := []map[string]interface{}{}
			err := client.Write(events)

 			Expect(err).To(BeNil())
			Expect(capturedRequest.URL.Path).To(Equal(custom_hec_endpoint))
		})
	})

	It("returns error on bad splunk host", func() {
		config.Host = ":"
		client := NewSplunk(config)
		events := []map[string]interface{}{}
		err := client.Write(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("protocol"))
	})

	It("Returns error on non-2xx response", func() {
		testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.WriteHeader(500)
			writer.Write([]byte("Internal server error"))
		}))

		config.Host = testServer.URL
		client := NewSplunk(config)
		events := []map[string]interface{}{}
		err := client.Write(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("500"))
	})

	It("Returns error from http client", func() {
		config.Host = "foo://example.com"
		client := NewSplunk(config)
		events := []map[string]interface{}{}
		err := client.Write(events)

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring("foo"))
	})
})
