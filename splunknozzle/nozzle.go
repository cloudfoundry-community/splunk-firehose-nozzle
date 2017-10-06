package splunknozzle

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/lager"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventrouter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/events"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsink"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsource"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventwriter"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"
	"github.com/google/uuid"
)

type SplunkFirehoseNozzle struct {
	config *Config
}

func NewSplunkFirehoseNozzle(config *Config) *SplunkFirehoseNozzle {
	return &SplunkFirehoseNozzle{
		config: config,
	}
}

// EventRouter creates EventRouter object and setup routes for interested events
func (s *SplunkFirehoseNozzle) EventRouter(cache cache.Cache, eventSink eventsink.Sink) (eventrouter.Router, error) {
	config := &eventrouter.Config{
		SelectedEvents: s.config.WantedEvents,
	}
	return eventrouter.New(cache, eventSink, config)
}

// PCFClient creates a client object which can talk to PCF
func (s *SplunkFirehoseNozzle) PCFClient() (*cfclient.Client, error) {
	cfConfig := &cfclient.Config{
		ApiAddress:        s.config.ApiEndpoint,
		Username:          s.config.User,
		Password:          s.config.Password,
		SkipSslValidation: s.config.SkipSSLCF,
	}

	return cfclient.NewClient(cfConfig)
}

// AppCache creates in-memory cache or boltDB cache
func (s *SplunkFirehoseNozzle) AppCache(client cache.AppClient, logger lager.Logger) (cache.Cache, error) {
	if s.config.AddAppInfo {
		c := cache.BoltdbConfig{
			Path:               s.config.BoltDBPath,
			IgnoreMissingApps:  s.config.IgnoreMissingApps,
			MissingAppCacheTTL: s.config.MissingAppCacheTTL,
			AppCacheTTL:        s.config.AppCacheTTL,
			Logger:             logger,
		}
		return cache.NewBoltdb(client, &c)
	}

	return cache.NewNoCache(), nil
}

// EventSink creates std sink or Splunk sink
func (s *SplunkFirehoseNozzle) EventSink(logger lager.Logger) (eventsink.Sink, error) {
	if s.config.Debug {
		return &eventsink.Std{}, nil
	}

	// EventWriter for writing events
	writerConfig := &eventwriter.SplunkConfig{
		Host:    			s.config.SplunkHost,
		Token:   			s.config.SplunkToken,
		Index:   			s.config.SplunkIndex,
		SkipSSL: 			s.config.SkipSSLSplunk,
		DisableKeepAlive:	s.config.HecKeepAliveOff,
		Logger:  logger,
	}

	var writers []eventwriter.Writer
	for i := 0; i < s.config.HecWorkers+1; i++ {
		splunkWriter := eventwriter.NewSplunk(writerConfig)
		writers = append(writers, splunkWriter)
	}

	parsedExtraFields, err := events.ParseExtraFields(s.config.ExtraFields)
	if err != nil {
		return nil, err
	}

	nozzleUUID := uuid.New().String()

	sinkConfig := &eventsink.SplunkConfig{
		FlushInterval:  s.config.FlushInterval,
		QueueSize:      s.config.QueueSize,
		BatchSize:      s.config.BatchSize,
		Retries:        s.config.Retries,
		Hostname:       s.config.JobHost,
		Version:        s.config.SplunkVersion,
		SubscriptionID: s.config.SubscriptionID,
		TraceLogging:   s.config.TraceLogging,
		ExtraFields:    parsedExtraFields,
		UUID:           nozzleUUID,
		Logger:         logger,
	}

	splunkSink := eventsink.NewSplunk(writers, sinkConfig)
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
		SkipSSL:        s.config.SkipSSLCF,
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
func (s *SplunkFirehoseNozzle) Run(shutdownChan chan os.Signal, logger lager.Logger) error {
	eventSink, err := s.EventSink(logger)
	if err != nil {
		return err
	}

	logger.Info("Running splunk-firehose-nozzle with following configuration variables ", s.config.ToMap())

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

	// Continuous Loop will run forever
	go func() {
		err := noz.Start()
		if err != nil {
			logger.Error("Firehose consumer exits with error", err)
		}
		shutdownChan <- os.Interrupt
	}()

	<-shutdownChan

	logger.Info("Splunk Nozzle is going to exit gracefully")
	noz.Close()
	return eventSink.Close()
}
