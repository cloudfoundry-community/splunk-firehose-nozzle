package auth_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"

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

			if response.code != 0 {
				writer.WriteHeader(response.code)
			}
			if response.body != nil {
				writer.Write(response.body)
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

		Context("client", func() {
			It("exist correctly calls endpoint", func() {
				responses = append(responses, testServerResponse{code: 404}, testServerResponse{code: 200})

				registrar.RegisterFirehoseClient("my-firehose-user", "my-firehose-secret")

				request := capturedRequests[0]
				Expect(request.request.Method).To(Equal("GET"))
				Expect(request.request.URL.Path).To(Equal("/oauth/clients/my-firehose-user"))
				Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
			})

			It("returns error when unable to determine if client exists", func() {
				responses = append(responses, testServerResponse{
					code: 301, //301 w/o location header forces error
				})

				err := registrar.RegisterFirehoseClient("my-firehose-user", "my-firehose-secret")

				Expect(err).NotTo(BeNil())
			})

			Context("client not present", func() {
				It("correctly calls create client", func() {
					responses = append(responses, testServerResponse{code: 404}, testServerResponse{code: 201})

					err := registrar.RegisterFirehoseClient("my-firehose-user", "my-firehose-secret")
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

					err := registrar.RegisterFirehoseClient("my-firehose-user", "my-firehose-secret")
					Expect(err).NotTo(BeNil())
				})
			})

			Context("client present", func() {
				It("correctly calls update client", func() {
					responses = append(responses, testServerResponse{code: 200}, testServerResponse{code: 200}, testServerResponse{code: 200})

					err := registrar.RegisterFirehoseClient("my-firehose-user", "my-firehose-secret")
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

					err := registrar.RegisterFirehoseClient("my-firehose-user", "my-firehose-secret")
					Expect(capturedRequests).To(HaveLen(2))
					Expect(err).NotTo(BeNil())
				})

				It("updates client secret", func() {
					responses = append(responses, testServerResponse{code: 200}, testServerResponse{code: 200}, testServerResponse{code: 200})

					err := registrar.RegisterFirehoseClient("my-firehose-user", "my-new-firehose-secret")
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

					err := registrar.RegisterFirehoseClient("my-firehose-user", "my-firehose-secret")
					Expect(capturedRequests).To(HaveLen(3))
					Expect(err).NotTo(BeNil())
				})
			})
		})

		Context("user", func() {
			setPasswordResponse := testServerResponse{code: 200}

			It("returns err when unable to list users", func() {
				responses = append(responses, testServerResponse{code: 500})

				_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("500"))
			})

			It("error when finding user responds with incorrect doe", func() {
				responses = append(responses, testServerResponse{code: 500})

				_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("500"))
			})

			Context("user exists", func() {
				BeforeEach(func() {
					responses = append(responses, testServerResponse{
						code: 200,
						body: []byte(`{
	"resources": [
		{"id": "5c2b3e19-bf76-441a-bfd9-6b499259d646"}
	],
	"startIndex": 1,
	"itemsPerPage": 100,
	"totalResults": 1
}`),
					})
				})

				It("correctly calls endpoint", func() {
					responses = append(responses, setPasswordResponse)

					_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).To(BeNil())

					request := capturedRequests[0]
					Expect(request.request.Method).To(Equal("GET"))
					Expect(request.request.URL.Path).To(Equal("/Users"))
					Expect(request.request.URL.RawQuery).To(Equal(`filter=userName+eq+"my-firehose-user"`))
					Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
				})

				It("correctly sets password", func() {
					responses = append(responses, setPasswordResponse)

					_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).To(BeNil())

					request := capturedRequests[1]
					Expect(request.request.Method).To(Equal("PUT"))
					Expect(request.request.URL.Path).To(Equal("/Users/5c2b3e19-bf76-441a-bfd9-6b499259d646/password"))
					Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
				})

				It("returns user id", func() {
					responses = append(responses, setPasswordResponse)

					id, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).To(BeNil())
					Expect(id).To(Equal("5c2b3e19-bf76-441a-bfd9-6b499259d646"))
				})

				It("returns error when set password returns incorrect code", func() {
					responses = append(responses, testServerResponse{code: 500})

					_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(ContainSubstring("500"))
				})
			})

			Context("user doesn't exist", func() {
				BeforeEach(func() {
					responses = append(responses, testServerResponse{
						code: 200,
						body: []byte(`{
	"resources": [],
	"startIndex": 1,
	"itemsPerPage": 100,
	"totalResults": 0
}`),
					})
				})

				It("correctly creates user", func() {
					responses = append(responses, testServerResponse{
						code: 201,
						body: []byte(`{
	"id": "6c840ccf-550d-489e-992c-629acd53500e"
}`),
					})
					responses = append(responses, setPasswordResponse)

					_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).To(BeNil())

					request := capturedRequests[1]
					Expect(request.request.Method).To(Equal("POST"))
					Expect(request.request.URL.Path).To(Equal("/Users"))
					Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
					Expect(request.request.Header.Get("Content-type")).To(Equal("application/json"))
				})

				It("error when create responds with incorrect status code", func() {
					responses = append(responses, testServerResponse{code: 500})

					_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(ContainSubstring("500"))
				})

				It("error when create responds with incorrect body", func() {
					responses = append(responses, testServerResponse{
						code: 201,
						body: []byte(`{"foo": "5c840ccfbar"}`),
					})
					responses = append(responses, setPasswordResponse)

					_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).NotTo(BeNil())
					Expect(err.Error()).To(ContainSubstring("foo"))
				})

				It("correctly sets password", func() {
					responses = append(responses, testServerResponse{
						code: 201,
						body: []byte(`{
	"id": "6c840ccf-550d-489e-992c-629acd53500e"
}`),
					}, setPasswordResponse)

					_, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).To(BeNil())

					request := capturedRequests[2]
					Expect(request.request.Method).To(Equal("PUT"))
					Expect(request.request.URL.Path).To(Equal("/Users/6c840ccf-550d-489e-992c-629acd53500e/password"))
					Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
				})

				It("returns user id", func() {
					responses = append(responses, testServerResponse{
						code: 201,
						body: []byte(`{
	"id": "6c840ccf-550d-489e-992c-629acd53500e"
}`),
					}, setPasswordResponse)

					id, err := registrar.RegisterUser("my-firehose-user", "my-firehose-password")
					Expect(err).To(BeNil())
					Expect(id).To(Equal("6c840ccf-550d-489e-992c-629acd53500e"))
				})
			})
		})

		Context("group", func() {
			var groupsResourceJson []byte

			BeforeEach(func() {
				groupsResourceJson = []byte(`{
    "resources": [{
            "meta": {
                "version": 4,
                "created": "2016-09-05T21:38:29.822Z",
                "lastModified": "2016-09-08T17:49:39.079Z"
            },
            "displayName": "cloud_controller.admin",
            "schemas": ["urn:scim:schemas:core:1.0"],
            "members": [{
                "origin": "uaa",
                "type": "USER",
                "value": "5c2b3e19-bf76-441a-bfd9-6b499259d646"
            }],
            "zoneId": "uaa",
            "id": "93aefed3-b30a-4a62-85b5-f3664b1dfe82"
        }
    ],
    "startIndex": 1,
    "totalResults": 1
}
`)
				r, _ := regexp.Compile(`\s`)
				groupsResourceJson = r.ReplaceAll(groupsResourceJson, []byte(""))
			})

			It("returns err when unable to list groups", func() {
				responses = append(responses, testServerResponse{code: 500})

				err := registrar.AddUserToGroup("user-id", "cloud_controller.admin")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("500"))
			})

			It("returns error if no group matching display name", func() {
				responses = append(responses, testServerResponse{
					code: 200,
					body: []byte(`{
	"resources": [],
	"totalResults": 0
}`),
				})

				err := registrar.AddUserToGroup("user-id", "non-existant-group")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("non-existant-group"))
			})

			It("error if json serialization doesn't match", func() {
				responses = append(responses, testServerResponse{
					code: 200,
					body: []byte(`{
    "resources": [{
            "meta": {
                "version": 4,
                "created": "2016-09-05T21:38:29.822Z",
                "lastModified": "2016-09-08T17:49:39.079Z"
            },
            "displayName": "cloud_controller.admin",
            "schemas": ["urn:scim:schemas:core:1.0"],
            "members": [{
				"origin": "uaa",
				"type": "USER",
				"value": "5c2b3e19-bf76-441a-bfd9-6b499259d646"
			}],
            "zoneId": "uaa",
            "id": "93aefed3-b30a-4a62-85b5-f3664b1dfe82",
            "key": "unexpected key"
        }
    ],
    "startIndex": 1,
    "totalResults": 1
}
`),
				})
				err := registrar.AddUserToGroup("user-id", "cloud_controller.admin")
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("schema"))
			})

			It("no action when already member", func() {
				responses = append(responses, testServerResponse{
					code: 200,
					body: groupsResourceJson,
				})
				err := registrar.AddUserToGroup("5c2b3e19-bf76-441a-bfd9-6b499259d646", "cloud_controller.admin")
				Expect(err).To(BeNil())

				Expect(capturedRequests).To(HaveLen(1))
			})

			It("correctly calls update group", func() {
				responses = append(responses,
					testServerResponse{
						code: 200,
						body: groupsResourceJson,
					},
					testServerResponse{code: 200},
				)
				err := registrar.AddUserToGroup("84587184-2698-4eb4-89d6-ab275ab4d01e", "cloud_controller.admin")
				Expect(err).To(BeNil())

				Expect(capturedRequests).To(HaveLen(2))

				request := capturedRequests[1]
				Expect(request.request.Method).To(Equal("PUT"))
				Expect(request.request.URL.Path).To(Equal("/Groups/93aefed3-b30a-4a62-85b5-f3664b1dfe82"))
				Expect(request.request.Header.Get("Authorization")).To(Equal("my-token"))
			})

			It("correct update group payload", func() {
				responses = append(responses,
					testServerResponse{
						code: 200,
						body: groupsResourceJson,
					},
					testServerResponse{code: 200},
				)
				err := registrar.AddUserToGroup("84587184-2698-4eb4-89d6-ab275ab4d01e", "cloud_controller.admin")
				Expect(err).To(BeNil())

				Expect(capturedRequests).To(HaveLen(2))

				request := capturedRequests[1]

				expectedPayload := []byte(`{
					"meta": {
						"version": 4,
						"created": "2016-09-05T21:38:29.822Z",
						"lastModified": "2016-09-08T17:49:39.079Z"
					},
					"displayName": "cloud_controller.admin",
					"schemas": ["urn:scim:schemas:core:1.0"],
					"members": [
						{
							"origin": "uaa",
							"type": "USER",
							"value": "5c2b3e19-bf76-441a-bfd9-6b499259d646"
						},
						{
							"origin": "uaa",
							"type": "USER",
							"value": "84587184-2698-4eb4-89d6-ab275ab4d01e"
						}
					],
					"zoneId": "uaa",
					"id": "93aefed3-b30a-4a62-85b5-f3664b1dfe82"
				}`)
				r, _ := regexp.Compile(`\s`)
				expectedPayload = r.ReplaceAll(expectedPayload, []byte(""))
				Expect(request.body).To(Equal(expectedPayload))
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
