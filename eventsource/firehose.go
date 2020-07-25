package eventsource

import (
	"crypto/tls"
	"errors"
	"time"

	"github.com/cloudfoundry/noaa/consumer"
	"github.com/cloudfoundry/sonde-go/events"
)

type FirehoseConfig struct {
	KeepAlive      time.Duration
	SkipSSL        bool
	Endpoint       string
	SubscriptionID string
}

type TokenClient interface {
	GetToken() (string, error)
}

type Firehose struct {
	config        *FirehoseConfig
	tokenClient   TokenClient
	eventConsumer *consumer.Consumer
}

func NewFirehose(tokenClient TokenClient, config *FirehoseConfig) *Firehose {
	c := consumer.New(config.Endpoint, &tls.Config{InsecureSkipVerify: config.SkipSSL}, nil)
	c.SetIdleTimeout(config.KeepAlive)

	f := &Firehose{
		config:        config,
		tokenClient:   tokenClient,
		eventConsumer: c,
	}
	c.RefreshTokenFrom(f)

	return f
}

func (f *Firehose) RefreshAuthToken() (string, error) {
	token, err := f.tokenClient.GetToken()
	if err != nil {
		return "", err
	}

	if token == "" {
		return "", errors.New("failed to refresh token")
	}

	return token, nil
}

func (f *Firehose) Open() error {
	return nil
}

func (f *Firehose) Close() error {
	return f.eventConsumer.Close()
}

func (f *Firehose) Read() (<-chan *events.Envelope, <-chan error) {
	return f.eventConsumer.Firehose(f.config.SubscriptionID, "")
}
