package uaago_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"

	"github.com/cloudfoundry-incubator/uaago"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	Context("GetOauthToken", func() {
		Context("with http", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in":599,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, err := client.GetAuthToken("myusername", "mypassword", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer good-token"))
			})
		})

		Context("with https", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in":599,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, err := client.GetAuthToken("myusername", "mypassword", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer good-token"))
			})
		})
	})

	Context("GetOauthToken With Expires_in", func() {
		Context("with http", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in":599,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token and expires_in from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, expiresIn, err := client.GetAuthTokenWithExpiresIn("myusername", "mypassword", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer good-token"))
				Expect(expiresIn).To(Equal(599))
			})
		})

		Context("with https", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in":598,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token and expires_in from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, expiresIn, err := client.GetAuthTokenWithExpiresIn("myusername", "mypassword", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer good-token"))
				Expect(expiresIn).To(Equal(598))
			})
		})

		Context("with invalid expires_in", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in":"invalid",
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token and expires_in from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, expiresIn, err := client.GetAuthTokenWithExpiresIn("myusername", "mypassword", true)
				Expect(err).To(HaveOccurred())
				Expect(token).To(Equal(""))
				Expect(expiresIn).To(Equal(-1))
			})
		})

		Context("without expires_in", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token missing the expires_in from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, expiresIn, err := client.GetAuthTokenWithExpiresIn("myusername", "mypassword", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer good-token"))
				Expect(expiresIn).To(Equal(0))
			})
		})
	})
})

func validRequest(request *http.Request) bool {
	isPost := request.Method == "POST"
	correctPath := request.URL.Path == "/oauth/token"
	correctType := request.Header.Get("content-type") == "application/x-www-form-urlencoded"
	request.ParseForm()
	hasClientId := len(request.PostForm.Get("client_id")) > 0
	hasGrantType := len(request.PostForm.Get("grant_type")) > 0

	return isPost && correctPath && correctType && hasClientId && hasGrantType
}
