package splunknozzle

import (
	"crypto/tls"
	"fmt"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"

	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/extrafields"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/auth"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/drain"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/sink"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
	"github.com/cloudfoundry/noaa/consumer"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/firehoseclient"
)

type SplunkFirehoseNozzle struct {
	c *Config
}

func NewSplunkFirehoseNozzle(config *Config) *SplunkFirehoseNozzle {
	return &SplunkFirehoseNozzle{
		c: config,
	}
}

// eventRouting creates eventRouting object and setup routings for interested events
func (s *SplunkFirehoseNozzle) eventRouting(cache caching.Caching, logClient logging.Logging) (*eventRouting.EventRouting, error) {
	events := eventRouting.NewEventRouting(cache, logClient)
	err := events.SetupEventRouting(s.c.WantedEvents)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// pcfClient creates a client object which can talk to PCF
func (s *SplunkFirehoseNozzle) pcfClient() (*cfclient.Client, error) {
	cfConfig := &cfclient.Config{
		ApiAddress:        s.c.ApiEndpoint,
		Username:          s.c.User,
		Password:          s.c.Password,
		SkipSslValidation: s.c.SkipSSL,
	}

	cfClient, err := cfclient.NewClient(cfConfig)
	if err != nil {
		return nil, err
	}
	return cfClient, nil
}

// appCache creates inmemory cache or boltDB cache
func (s *SplunkFirehoseNozzle) appCache(cfClient *cfclient.Client) caching.Caching {
	if s.c.AddAppInfo {
		cache := caching.NewCachingBolt(cfClient, s.c.BoltDBPath)
		cache.CreateBucket()
		return cache
	}

	return caching.NewCachingEmpty()
}

// logClient creates std logging or Splunk logging
func (s *SplunkFirehoseNozzle) logClient(logger lager.Logger) (logging.Logging, error) {
	if s.c.Debug {
		return &drain.LoggingStd{}, nil
	}

	parsedExtraFields, err := extrafields.ParseExtraFields(s.c.ExtraFields)
	if err != nil {
		return nil, err
	}

	// SplunkClient for nozzle internal logging
	splunkClient := splunk.NewSplunkClient(s.c.SplunkToken, s.c.SplunkHost, s.c.SplunkIndex, parsedExtraFields, s.c.SkipSSL, logger)
	logger.RegisterSink(sink.NewSplunkSink(s.c.JobName, s.c.JobIndex, s.c.JobHost, splunkClient))

	// SplunkClients for raw event POST
	var splunkClients []splunk.SplunkClient
	for i := 0; i < s.c.HecWorkers; i++ {
		splunkClient := splunk.NewSplunkClient(s.c.SplunkToken, s.c.SplunkHost, s.c.SplunkIndex, parsedExtraFields, s.c.SkipSSL, logger)
		splunkClients = append(splunkClients, splunkClient)
	}

	loggingConfig := &drain.LoggingConfig{
		FlushInterval: s.c.FlushInterval,
		QueueSize:     s.c.QueueSize,
		BatchSize:     s.c.BatchSize,
		Retries:       s.c.Retries,
	}

	splunkLog := drain.NewLoggingSplunk(logger, splunkClients, loggingConfig)
	if !splunkLog.Connect() {
		return nil, fmt.Errorf("failed to connect splunk")
	}
	return splunkLog, nil
}

// firehoseConsumer creates firehose Consumer object which can subscribes the PCF firehose events
func (s *SplunkFirehoseNozzle) firehoseConsumer(pcfClient *cfclient.Client) *consumer.Consumer {
	dopplerEndpoint := pcfClient.Endpoint.DopplerEndpoint
	tokenRefresher := auth.NewTokenRefreshAdapter(pcfClient)
	consumer := consumer.New(dopplerEndpoint, &tls.Config{InsecureSkipVerify: s.c.SkipSSL}, nil)
	consumer.RefreshTokenFrom(tokenRefresher)
	consumer.SetIdleTimeout(s.c.KeepAlive)
	return consumer
}

// firehoseClient creates FirehoseNozzle object which glues the event source and event sink
func (s *SplunkFirehoseNozzle) firehoseClient(consumer *consumer.Consumer, events *eventRouting.EventRouting) *firehoseclient.FirehoseNozzle {
	firehoseConfig := &firehoseclient.FirehoseConfig{
		FirehoseSubscriptionID: s.c.SubscriptionID,
	}

	return firehoseclient.NewFirehoseNozzle(consumer, events, firehoseConfig)
}

// Run creates all necessary objects, reading events from PCF firehose and sending to target Splunk index
// It runs forever until something goes wrong
func (s *SplunkFirehoseNozzle) Run(logger lager.Logger) error {
	logClient, err := s.logClient(logger)
	if err != nil {
		return err
	}

	params := lager.Data{
		"version":        s.c.Version,
		"branch":         s.c.Branch,
		"commit":         s.c.Commit,
		"buildos":        s.c.BuildOS,
		"flush-interval": s.c.FlushInterval,
		"queue-size":     s.c.QueueSize,
		"batch-size":     s.c.BatchSize,
		"workers":        s.c.HecWorkers,
	}
	logger.Info("splunk-firehose-nozzle runs", params)

	pcfClient, err := s.pcfClient()
	if err != nil {
		return err
	}

	appCache := s.appCache(pcfClient)

	events, err := s.eventRouting(appCache, logClient)
	if err != nil {
		return err
	}

	consumer := s.firehoseConsumer(pcfClient)
	firehoseClient := s.firehoseClient(consumer, events)
	if err := firehoseClient.Start(); err != nil {
		return err
	}
	return nil
}
