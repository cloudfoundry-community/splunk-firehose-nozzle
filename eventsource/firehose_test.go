package eventsource_test

import (
	"errors"
	"time"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsource"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/testing"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Eventsource", func() {
	var (
		config *FirehoseConfig
	)

	BeforeEach(func() {
		config = &FirehoseConfig{
			KeepAlive:      25 * time.Second,
			SkipSSL:        true,
			Endpoint:       "ws://localhost:9912",
			SubscriptionID: "testing",
		}
	})

	It("refresh requests and returns token", func() {
		called := false
		tokenClient := &testing.TokenClientMock{
			GetTokenFn: func() (string, error) {
				called = true
				return "my-token", nil
			},
		}

		f := NewFirehose(tokenClient, config)
		token, err := f.RefreshAuthToken()
		Expect(err).To(BeNil())
		Expect(called).To(BeTrue())
		Expect(token).To(Equal("my-token"))
	})

	It("returns error when no token", func() {
		tokenClient := &testing.TokenClientMock{
			GetTokenFn: func() (string, error) {
				return "", nil
			},
		}

		f := NewFirehose(tokenClient, config)
		token, err := f.RefreshAuthToken()
		Expect(err).NotTo(BeNil())
		Expect(token).To(Equal(""))
	})

	It("returns getToken's error", func() {
		tokenClient := &testing.TokenClientMock{
			GetTokenFn: func() (string, error) {
				return "", errors.New("Failed to get token")
			},
		}

		f := NewFirehose(tokenClient, config)
		token, err := f.RefreshAuthToken()

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(Equal("Failed to get token"))
		Expect(token).To(Equal(""))
	})

	It("Open", func() {
		tokenClient := &testing.TokenClientMock{
			GetTokenFn: func() (string, error) {
				return "token", nil
			},
		}

		f := NewFirehose(tokenClient, config)
		err := f.Open()
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("Read", func() {
		tokenClient := &testing.TokenClientMock{
			GetTokenFn: func() (string, error) {
				return "token", nil
			},
		}

		f := NewFirehose(tokenClient, config)
		eventChan, errChan := f.Read()

		var e *events.Envelope
		var err error

		select {
		case e = <-eventChan:
		case err = <-errChan:
		}

		Expect(e).To(BeNil())
		Ω(err).Should(HaveOccurred())
	})

	It("close", func() {
		tokenClient := &testing.TokenClientMock{
			GetTokenFn: func() (string, error) {
				return "token", nil
			},
		}

		f := NewFirehose(tokenClient, config)
		err := f.Close()
		Ω(err).Should(HaveOccurred())
	})

})
