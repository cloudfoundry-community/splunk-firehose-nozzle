package auth_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"code.cloudfoundry.org/lager"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/auth"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("uaa_registrar", func() {
	type testServerResponse struct {
		body []byte
		code int
	}

	type testServerRequest struct {
		request *http.Request
		body    []byte
	}

	var (
		testServer       *httptest.Server
		capturedRequests []*testServerRequest
		responses        []testServerResponse
		logger           lager.Logger

		tokenRefresher *MockTokenRefresher
	)

	BeforeEach(func() {
		capturedRequests = []*testServerRequest{}
		responses = []testServerResponse{}

		testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			requestBody, err := ioutil.ReadAll(request.Body)
			if err != nil {
				panic(err)
			}
			capturedRequests = append(capturedRequests, &testServerRequest{
				request: request,
				body:    requestBody,
			})

			response := responses[0]
			responses = responses[1:]

			if response.body != nil {
				writer.Write(response.body)
			}
			if response.code != 200 {
				writer.WriteHeader(response.code)
			}

		}))

		logger = lager.NewLogger("test")
		tokenRefresher = &MockTokenRefresher{}
	})

	AfterEach(func() {
		testServer.Close()
	})

	It("new should fetch auth token", func() {
		called := false
		tokenRefresher.RefreshAuthTokenFn = func() (string, error) {
			called = true
			return "my-token", nil
		}

		_, err := NewUaaRegistrar(
			"https://uaa.example.com", tokenRefresher, true, logger,
		)

		Expect(err).To(BeNil())
		Expect(called).To(BeTrue())
	})

	It("new should return error", func() {
		tokenRefresher.RefreshAuthTokenFn = func() (string, error) {
			return "", errors.New("some error")
		}

		registrar, err := NewUaaRegistrar(
			testServer.URL, tokenRefresher, true, logger,
		)

		Expect(registrar).To(BeNil())
		Expect(err).To(Equal(errors.New("some error")))
	})

	Context("with registrar", func() {
		var registrar UaaRegistrar

		BeforeEach(func() {
			registrar, _ = NewUaaRegistrar(
				testServer.URL, tokenRefresher, true, logger,
			)
		})

		It("exist correctly calls endpoint", func() {
			responses = append(responses, testServerResponse{code: 404}, testServerResponse{code: 200})

			registrar.RegisterFirehose("my-firehose-user", "my-firehose-secret")

			request := capturedRequests[0]
			Expect(request.request.Method).To(Equal("GET"))
			Expect(request.request.URL.Path).To(Equal("/oauth/clients/my-firehose-user"))
			Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
		})

		It("returns error when unable to determine if client exists", func() {
			responses = append(responses, testServerResponse{
				code: 301, //301 w/o location header forces error
			})

			err := registrar.RegisterFirehose("my-firehose-user", "my-firehose-secret")

			Expect(err).NotTo(BeNil())
		})

		Context("client not present", func() {
			It("correctly calls create client", func() {
				responses = append(responses, testServerResponse{code: 404}, testServerResponse{code: 201})

				err := registrar.RegisterFirehose("my-firehose-user", "my-firehose-secret")
				Expect(err).To(BeNil())

				request := capturedRequests[1]
				Expect(request.request.Method).To(Equal("POST"))
				Expect(request.request.URL.Path).To(Equal("/oauth/clients"))
				Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
				Expect(request.request.Header.Get("Content-type")).To(Equal("application/json"))

				var payload map[string]interface{}
				err = json.Unmarshal(request.body, &payload)
				Expect(err).To(BeNil())

				Expect(payload["client_id"]).To(Equal("my-firehose-user"))
				Expect(payload["client_secret"]).To(Equal("my-firehose-secret"))
				Expect(payload["scope"]).To(Equal([]interface{}{"openid", "oauth.approvals", "doppler.firehose"}))
				Expect(payload["authorized_grant_types"]).To(Equal([]interface{}{"client_credentials"}))
			})

			It("returns error if create client fails", func() {
				responses = append(responses, testServerResponse{code: 404}, testServerResponse{code: 500})

				err := registrar.RegisterFirehose("my-firehose-user", "my-firehose-secret")
				Expect(err).NotTo(BeNil())
			})
		})

		Context("client present", func() {
			It("correctly calls update client", func() {
				responses = append(responses, testServerResponse{code: 200}, testServerResponse{code: 200}, testServerResponse{code: 200})

				err := registrar.RegisterFirehose("my-firehose-user", "my-firehose-secret")
				Expect(err).To(BeNil())
				Expect(capturedRequests).To(HaveLen(3))

				request := capturedRequests[1]
				Expect(request.request.Method).To(Equal("PUT"))
				Expect(request.request.URL.Path).To(Equal("/oauth/clients/my-firehose-user"))
				Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
				Expect(request.request.Header.Get("Content-type")).To(Equal("application/json"))

				var payload map[string]interface{}
				err = json.Unmarshal(request.body, &payload)
				Expect(err).To(BeNil())

				Expect(payload["client_id"]).To(Equal("my-firehose-user"))
				Expect(payload["scope"]).To(Equal([]interface{}{"openid", "oauth.approvals", "doppler.firehose"}))
				Expect(payload["authorized_grant_types"]).To(Equal([]interface{}{"client_credentials"}))
			})

			It("returns error if update client fails", func() {
				responses = append(responses, testServerResponse{code: 200}, testServerResponse{code: 500})

				err := registrar.RegisterFirehose("my-firehose-user", "my-firehose-secret")
				Expect(capturedRequests).To(HaveLen(2))
				Expect(err).NotTo(BeNil())
			})

			It("updates client secret", func() {
				responses = append(responses, testServerResponse{code: 200}, testServerResponse{code: 200}, testServerResponse{code: 200})

				err := registrar.RegisterFirehose("my-firehose-user", "my-new-firehose-secret")
				Expect(err).To(BeNil())
				Expect(capturedRequests).To(HaveLen(3))

				request := capturedRequests[2]
				Expect(request.request.Method).To(Equal("PUT"))
				Expect(request.request.URL.Path).To(Equal("/oauth/clients/my-firehose-user/secret"))
				Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
				Expect(request.request.Header.Get("Content-type")).To(Equal("application/json"))

				var payload map[string]interface{}
				err = json.Unmarshal(request.body, &payload)
				Expect(err).To(BeNil())

				Expect(payload["secret"]).To(Equal("my-new-firehose-secret"))
			})

			It("returns error if update client secret fails", func() {
				responses = append(responses, testServerResponse{code: 200}, testServerResponse{code: 200}, testServerResponse{code: 500})

				err := registrar.RegisterFirehose("my-firehose-user", "my-firehose-secret")
				Expect(capturedRequests).To(HaveLen(3))
				Expect(err).NotTo(BeNil())
			})
		})
	})
})

type MockTokenRefresher struct {
	RefreshAuthTokenFn func() (string, error)
}

func (m *MockTokenRefresher) RefreshAuthToken() (string, error) {
	if m.RefreshAuthTokenFn != nil {
		return m.RefreshAuthTokenFn()
	}
	return "my-token", nil
}
