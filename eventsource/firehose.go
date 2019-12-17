package eventsource

import (
	"code.cloudfoundry.org/go-loggregator"
	"github.com/cloudfoundry/sonde-go/events"
	"net/http"
	"time"
)

// FirehoseConfig struct with 4 fields of different types.
type FirehoseConfig struct {
	KeepAlive      time.Duration
	SkipSSL        bool
	Endpoint       string
	SubscriptionID string
}

// Doer is used to make HTTP requests to the RLP Gateway.
type doer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Firehose struct with fields of types FirehoseConfig, doer and V2adapter.
type Firehose struct {
	config      *FirehoseConfig
	tokenClient doer
	v2          V2Adapter
}

// NewFirehose the HTTP client.
func NewFirehose(tokenClient doer, config *FirehoseConfig) *Firehose {
	c := loggregator.NewRLPGatewayClient(
		config.Endpoint,
		loggregator.WithRLPGatewayHTTPClient(tokenClient),
	)

	f := &Firehose{
		config:      config,
		tokenClient: tokenClient,
		v2:          NewV2Adapter(c),
	}

	return f
}

// Open initiates Firehose
func (f *Firehose) Open() error {
	return nil
}

// Close closes Firehose
func (f *Firehose) Close() error {
	return nil
}

// Read reads envelope stream
func (f *Firehose) Read() <-chan *events.Envelope {
	return f.v2.Firehose(f.config.SubscriptionID)
}
