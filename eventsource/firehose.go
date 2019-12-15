package eventsource

import (
	//"crypto/tls"
	//"errors"

	"code.cloudfoundry.org/go-loggregator"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
	"net/http"
	"time"
)

type FirehoseConfig struct {
	KeepAlive      time.Duration
	SkipSSL        bool
	Endpoint       string
	SubscriptionID string
}

//type TokenClient interface {
//	GetToken() (string, error)
//}

type doer interface {
	Do(req *http.Request) (*http.Response, error)
}
type Firehose struct {
	config        *FirehoseConfig
	tokenClient   doer
	eventConsumer *consumer.Consumer
	a             V2Adapter
	stopReading   chan struct{}
	stopRouting   chan struct{}
}

func NewFirehose(tokenClient doer, config *FirehoseConfig) *Firehose {
	c := loggregator.NewRLPGatewayClient(
		config.Endpoint,
		loggregator.WithRLPGatewayHTTPClient(tokenClient),
	)
	//c.SetIdleTimeout(config.KeepAlive)

	f := &Firehose{
		config:      config,
		tokenClient: tokenClient,
		a:           NewV2Adapter(c),
		stopReading: make(chan struct{}),
		stopRouting: make(chan struct{}),
	}
	//c.RefreshTokenFrom(f)

	return f
}

//
//func (f *Firehose) RefreshAuthToken() (string, error) {
//	token, err := f.tokenClient.GetToken()
//	if err != nil {
//		return "", err
//	}
//
//	if token == "" {
//		return "", errors.New("failed to refresh token")
//	}
//
//	return token, nil
//}

func (f *Firehose) Open() error {
	return nil
}

func (f *Firehose) Close() error {
	//return f.eventConsumer.Close()
	return nil
}

func (f *Firehose) Read() <-chan *events.Envelope {
	return f.a.Firehose(f.config.SubscriptionID)
}
