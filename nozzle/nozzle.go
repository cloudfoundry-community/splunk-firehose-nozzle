package splunknozzle

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/lager"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/auth"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/caching"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/drain"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventRouting"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/extrafields"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/logging"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/sink"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
	"github.com/cloudfoundry/noaa/consumer"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/firehoseclient"
)

type SplunkFirehoseNozzle struct {
	config *Config
}

func NewSplunkFirehoseNozzle(config *Config) *SplunkFirehoseNozzle {
	return &SplunkFirehoseNozzle{
		config: config,
	}
}

// eventRouting creates eventRouting object and setup routings for interested events
func (s *SplunkFirehoseNozzle) eventRouting(cache caching.Caching, logClient logging.Logging) (eventRouting.EventRouting, error) {
	events := eventRouting.NewEventRouting(cache, logClient)
	err := events.SetupEventRouting(s.config.WantedEvents)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// pcfClient creates a client object which can talk to PCF
func (s *SplunkFirehoseNozzle) pcfClient() (*cfclient.Client, error) {
	cfConfig := &cfclient.Config{
		ApiAddress:        s.config.ApiEndpoint,
		Username:          s.config.User,
		Password:          s.config.Password,
		SkipSslValidation: s.config.SkipSSL,
	}

	cfClient, err := cfclient.NewClient(cfConfig)
	if err != nil {
		return nil, err
	}
	return cfClient, nil
}

// appCache creates inmemory cache or boltDB cache
func (s *SplunkFirehoseNozzle) appCache(cfClient *cfclient.Client) (caching.Caching, error) {
	if s.config.AddAppInfo {
		c := caching.CachingBoltConfig{
			Path:               s.config.BoltDBPath,
			IgnoreMissingApps:  s.config.IgnoreMissingApps,
			MissingAppCacheTTL: s.config.MissingAppCacheTTL,
			AppCacheTTL:        s.config.AppCacheTTL,
		}
		return caching.NewCachingBolt(cfClient, &c)
	}

	return caching.NewCachingEmpty(), nil
}

// logClient creates std logging or Splunk logging
func (s *SplunkFirehoseNozzle) logClient(logger lager.Logger) (logging.Logging, error) {
	if s.config.Debug {
		return &drain.LoggingStd{}, nil
	}

	parsedExtraFields, err := extrafields.ParseExtraFields(s.config.ExtraFields)
	if err != nil {
		return nil, err
	}

	// SplunkClient for nozzle internal logging
	splunkClient := splunk.NewSplunkClient(s.config.SplunkToken, s.config.SplunkHost, s.config.SplunkIndex, parsedExtraFields, s.config.SkipSSL, logger)
	logger.RegisterSink(sink.NewSplunkSink(s.config.JobName, s.config.JobIndex, s.config.JobHost, splunkClient))

	// SplunkClients for raw event POST
	var splunkClients []splunk.SplunkClient
	for i := 0; i < s.config.HecWorkers; i++ {
		splunkClient := splunk.NewSplunkClient(s.config.SplunkToken, s.config.SplunkHost, s.config.SplunkIndex, parsedExtraFields, s.config.SkipSSL, logger)
		splunkClients = append(splunkClients, splunkClient)
	}

	loggingConfig := &drain.LoggingConfig{
		FlushInterval: s.config.FlushInterval,
		QueueSize:     s.config.QueueSize,
		BatchSize:     s.config.BatchSize,
		Retries:       s.config.Retries,
	}

	splunkLog := drain.NewLoggingSplunk(logger, splunkClients, loggingConfig)
	if err := splunkLog.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect splunk")
	}
	return splunkLog, nil
}

// firehoseConsumer creates firehose Consumer object which can subscribes the PCF firehose events
func (s *SplunkFirehoseNozzle) firehoseConsumer(pcfClient *cfclient.Client) *consumer.Consumer {
	dopplerEndpoint := pcfClient.Endpoint.DopplerEndpoint
	tokenRefresher := auth.NewTokenRefreshAdapter(pcfClient)
	c := consumer.New(dopplerEndpoint, &tls.Config{InsecureSkipVerify: s.config.SkipSSL}, nil)
	c.RefreshTokenFrom(tokenRefresher)
	c.SetIdleTimeout(s.config.KeepAlive)
	return c
}

// firehoseClient creates FirehoseNozzle object which glues the event source and event sink
func (s *SplunkFirehoseNozzle) firehoseClient(consumer *consumer.Consumer, events eventRouting.EventRouting, logger lager.Logger) *firehoseclient.FirehoseNozzle {
	firehoseConfig := &firehoseclient.FirehoseConfig{
		Logger: logger,

		FirehoseSubscriptionID: s.config.SubscriptionID,
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
		"version": s.config.Version,
		"branch":  s.config.Branch,
		"commit":  s.config.Commit,
		"buildos": s.config.BuildOS,

		"add-app-info":             s.config.AddAppInfo,
		"ignore-missing-app":       s.config.IgnoreMissingApps,
		"app-cache-invalidate-ttl": s.config.AppCacheTTL,

		"flush-interval": s.config.FlushInterval,
		"queue-size":     s.config.QueueSize,
		"batch-size":     s.config.BatchSize,
		"workers":        s.config.HecWorkers,
	}
	logger.Info("splunk-firehose-nozzle runs", params)

	pcfClient, err := s.pcfClient()
	if err != nil {
		return err
	}

	appCache, err := s.appCache(pcfClient)
	if err != nil {
		return err
	}

	err = appCache.Open()
	if err != nil {
		return err
	}
	defer appCache.Close()

	eventRouter, err := s.eventRouting(appCache, logClient)
	if err != nil {
		return err
	}

	c := s.firehoseConsumer(pcfClient)
	firehoseClient := s.firehoseClient(c, eventRouter, logger)

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := firehoseClient.Start()
		if err != nil {
			logger.Error("Firehose consumer exits with error", err)
		}
		shutdown <- os.Interrupt
	}()

	<-shutdown

	logger.Info("Splunk Nozzle is going to exit gracefully")
	firehoseClient.Close()
	return logClient.Close()
}
