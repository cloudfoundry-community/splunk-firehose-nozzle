package auth_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/testing"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/auth"
)

var _ = Describe("uaa_token_fetcher", func() {
	It("refresh requests and returns token", func() {
		called := false
		tokenGetter := &testing.MockTokenGetter{
			GetTokenFn: func() (string, error) {
				called = true
				return "my-token", nil
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
			GetTokenFn: func() (string, error) {
				return "", nil
			},
		}

		fetcher := NewTokenRefreshAdapter(tokenGetter)
		token, err := fetcher.RefreshAuthToken()

		Expect(err).NotTo(BeNil())
		Expect(token).To(Equal(""))
	})

	It("returns getToken's error", func() {
		tokenGetter := &testing.MockTokenGetter{
			GetTokenFn: func() (string, error) {
				return "", errors.New("Failed to get token")
			},
		}

		fetcher := NewTokenRefreshAdapter(tokenGetter)
		token, err := fetcher.RefreshAuthToken()

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("Failed to get token"))
		Expect(token).To(Equal(""))
	})
})
