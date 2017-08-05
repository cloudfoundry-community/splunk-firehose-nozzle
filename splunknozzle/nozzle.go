package splunknozzle

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/lager"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventrouter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/events"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsink"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsource"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventwriter"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"
)

type SplunkFirehoseNozzle struct {
	config *Config
}

func NewSplunkFirehoseNozzle(config *Config) *SplunkFirehoseNozzle {
	return &SplunkFirehoseNozzle{
		config: config,
	}
}

// EventRouter creates EventRouter object and setup routings for interested events
func (s *SplunkFirehoseNozzle) EventRouter(cache cache.Cache, eventSink eventsink.Sink) (eventrouter.Router, error) {
	config := &eventrouter.Config{
		SelectedEvents: s.config.WantedEvents,
		ExtraFields:    s.config.ExtraFields,
	}
	return eventrouter.New(cache, eventSink, config)
}

// PCFClient creates a client object which can talk to PCF
func (s *SplunkFirehoseNozzle) PCFClient() (*cfclient.Client, error) {
	cfConfig := &cfclient.Config{
		ApiAddress:        s.config.ApiEndpoint,
		Username:          s.config.User,
		Password:          s.config.Password,
		SkipSslValidation: s.config.SkipSSL,
	}

	return cfclient.NewClient(cfConfig)
}

// AppCache creates inmemory cache or boltDB cache
func (s *SplunkFirehoseNozzle) AppCache(cfClient *cfclient.Client, logger lager.Logger) (cache.Cache, error) {
	if s.config.AddAppInfo {
		c := cache.BoltdbCacheConfig{
			Path:               s.config.BoltDBPath,
			IgnoreMissingApps:  s.config.IgnoreMissingApps,
			MissingAppCacheTTL: s.config.MissingAppCacheTTL,
			AppCacheTTL:        s.config.AppCacheTTL,
			Logger:             logger,
		}
		return cache.NewBoltdbCache(cfClient, &c)
	}

	return cache.NewNoCache(), nil
}

// EventSink creates std sink or Splunk sink
func (s *SplunkFirehoseNozzle) EventSink(logger lager.Logger) (eventsink.Sink, error) {
	if s.config.Debug {
		return &eventsink.Std{}, nil
	}

	parsedExtraFields, err := events.ParseExtraFields(s.config.ExtraFields)
	if err != nil {
		return nil, err
	}

	// EventWriter for writing events
	var writers []eventwriter.Writer
	for i := 0; i < s.config.HecWorkers+1; i++ {
		writer := eventwriter.NewSplunk(s.config.SplunkToken, s.config.SplunkHost, s.config.SplunkIndex, parsedExtraFields, s.config.SkipSSL, logger)
		writers = append(writers, writer)
	}

	config := &eventsink.SplunkConfig{
		FlushInterval: s.config.FlushInterval,
		QueueSize:     s.config.QueueSize,
		BatchSize:     s.config.BatchSize,
		Retries:       s.config.Retries,
		Hostname:      s.config.JobHost,
		Logger:        logger,
	}

	splunkSink := eventsink.NewSplunk(writers, config)
	if err := splunkSink.Open(); err != nil {
		return nil, fmt.Errorf("failed to connect splunk")
	}

	logger.RegisterSink(splunkSink)
	return splunkSink, nil
}

// EventSource creates eventsource.Source object which can read events from
func (s *SplunkFirehoseNozzle) EventSource(pcfClient *cfclient.Client) *eventsource.Firehose {
	config := &eventsource.FirehoseConfig{
		KeepAlive:      s.config.KeepAlive,
		SkipSSL:        s.config.SkipSSL,
		Endpoint:       pcfClient.Endpoint.DopplerEndpoint,
		SubscriptionID: s.config.SubscriptionID,
	}

	return eventsource.NewFirehose(pcfClient, config)
}

// Nozzle creates a Nozzle object which glues the event source and event router
func (s *SplunkFirehoseNozzle) Nozzle(eventSource eventsource.Source, eventRouter eventrouter.Router, logger lager.Logger) *nozzle.Nozzle {
	firehoseConfig := &nozzle.Config{
		Logger: logger,
	}

	return nozzle.New(eventSource, eventRouter, firehoseConfig)
}

// Run creates all necessary objects, reading events from PCF firehose and sending to target Splunk index
// It runs forever until something goes wrong
func (s *SplunkFirehoseNozzle) Run(logger lager.Logger) error {
	eventSink, err := s.EventSink(logger)
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

	pcfClient, err := s.PCFClient()
	if err != nil {
		return err
	}

	appCache, err := s.AppCache(pcfClient, logger)
	if err != nil {
		return err
	}

	err = appCache.Open()
	if err != nil {
		return err
	}
	defer appCache.Close()

	eventRouter, err := s.EventRouter(appCache, eventSink)
	if err != nil {
		return err
	}

	eventSource := s.EventSource(pcfClient)
	noz := s.Nozzle(eventSource, eventRouter, logger)

	shutdown := make(chan os.Signal, 2)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := noz.Start()
		if err != nil {
			logger.Error("Firehose consumer exits with error", err)
		}
		shutdown <- os.Interrupt
	}()

	<-shutdown

	logger.Info("Splunk Nozzle is going to exit gracefully")
	noz.Close()
	return eventSink.Close()
}
