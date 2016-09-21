package auth_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/testing"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/auth"
)

var _ = Describe("uaa_token_fetcher", func() {
	It("refresh requests and returns token", func() {
		called := false
		tokenGetter := &testing.MockTokenGetter{
			GetTokenFn: func() string {
				called = true
				return "my-token"
			},
		}

		fetcher := NewTokenRefreshAdapter(tokenGetter)
		token, err := fetcher.RefreshAuthToken()

		Expect(err).To(BeNil())

		Expect(called).To(BeTrue())
		Expect(token).To(Equal("my-token"))
	})

	It("returns error when no token", func() {
		tokenGetter := &testing.MockTokenGetter{
			GetTokenFn: func() string {
				//cfclient.client swallows error and returns empty, conform to that failure mode
				return ""
			},
		}

		fetcher := NewTokenRefreshAdapter(tokenGetter)
		token, err := fetcher.RefreshAuthToken()

		Expect(err).NotTo(BeNil())
		Expect(token).To(Equal(""))
	})
})
