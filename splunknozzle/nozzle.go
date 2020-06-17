package splunknozzle

import (
	"code.cloudfoundry.org/lager"
	"fmt"
	"log"
	"os"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventrouter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/events"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsink"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventsource"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventwriter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"
	"github.com/cloudfoundry-incubator/uaago"
	"github.com/google/uuid"
	"strings"
)

var gatewayLoggerAddr = log.New(os.Stderr, "RLP_Gateway Error - ", 3)
var gatewayErrChan = make(chan error, 1)

// SplunkFirehoseNozzle struct type with config fields.
type SplunkFirehoseNozzle struct {
	config *Config
	logger lager.Logger
}

// NewSplunkFirehoseNozzle create new function of type *SplunkFirehoseNozzle
func NewSplunkFirehoseNozzle(config *Config, logger lager.Logger) *SplunkFirehoseNozzle {
	return &SplunkFirehoseNozzle{
		config: config,
		logger: logger,
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
		ClientID:          s.config.ClientID,
		ClientSecret:      s.config.ClientSecret,
	}

	return cfclient.NewClient(cfConfig)
}

// AppCache creates in-memory cache or boltDB cache
func (s *SplunkFirehoseNozzle) AppCache(client cache.AppClient) (cache.Cache, error) {
	if s.config.AddAppInfo {
		c := cache.BoltdbConfig{
			Path:               s.config.BoltDBPath,
			IgnoreMissingApps:  s.config.IgnoreMissingApps,
			MissingAppCacheTTL: s.config.MissingAppCacheTTL,
			AppCacheTTL:        s.config.AppCacheTTL,
			OrgSpaceCacheTTL:   s.config.OrgSpaceCacheTTL,
			Logger:             s.logger,
		}
		return cache.NewBoltdb(client, &c)
	}

	return cache.NewNoCache(), nil
}

// EventSink creates std sink or Splunk sink
func (s *SplunkFirehoseNozzle) EventSink() (eventsink.Sink, error) {
	if s.config.Debug {
		return &eventsink.Std{}, nil
	}
	// EventWriter for writing events
	writerConfig := &eventwriter.SplunkConfig{
		Host:    s.config.SplunkHost,
		Token:   s.config.SplunkToken,
		Index:   s.config.SplunkIndex,
		SkipSSL: s.config.SkipSSLSplunk,
		Logger:  s.logger,
	}

	var writers []eventwriter.Writer
	for i := 0; i < s.config.HecWorkers+1; i++ {
		splunkWriter := eventwriter.NewSplunk(writerConfig)
		writers = append(writers, splunkWriter)
	}

	parsedExtraFields, err := events.ParseExtraFields(s.config.ExtraFields)
	if err != nil {
		s.logger.Error("Error at parsing extra fields", nil)
		return nil, err
	}

	nozzleUUID := uuid.New().String()

	sinkConfig := &eventsink.SplunkConfig{
		FlushInterval:         s.config.FlushInterval,
		QueueSize:             s.config.QueueSize,
		BatchSize:             s.config.BatchSize,
		Retries:               s.config.Retries,
		Hostname:              s.config.JobHost,
		Version:               s.config.SplunkVersion,
		SubscriptionID:        s.config.SubscriptionID,
		TraceLogging:          s.config.TraceLogging,
		ExtraFields:           parsedExtraFields,
		UUID:                  nozzleUUID,
		Logger:                s.logger,
		StatusMonitorInterval: s.config.StatusMonitorInterval,
	}

	splunkSink := eventsink.NewSplunk(writers, sinkConfig)
	splunkSink.Open()

	s.logger.RegisterSink(splunkSink)
	if s.config.StatusMonitorInterval > time.Second*0 {
		go splunkSink.LogStatus()
	}
	return splunkSink, nil
}

// EventSource creates eventsource.Source object which can read events from
func (s *SplunkFirehoseNozzle) EventSource(pcfClient *cfclient.Client) (*eventsource.Firehose, error) {
	config := &eventsource.FirehoseConfig{
		KeepAlive:             s.config.KeepAlive,
		SkipSSL:               s.config.SkipSSLCF,
		Endpoint:              strings.Replace(s.config.ApiEndpoint, "api", "log-stream", 1),
		SubscriptionID:        s.config.SubscriptionID,
		GatewayLoggerAddr:     gatewayLoggerAddr,
		GatewayErrChanAddr:    &gatewayErrChan,
		GatewayMaxRetries:     s.config.RLPGatewayRetries,
		StatusMonitorInterval: s.config.StatusMonitorInterval,
		Logger:                s.logger,
	}
	uaa, err := uaago.NewClient(pcfClient.Endpoint.AuthEndpoint)
	if err != nil {
		fmt.Println("unable to connect to get token from uaa", err)
		return nil, err
	}

	ac := eventsource.NewHttp(uaa, pcfClient.Config.ClientID, pcfClient.Config.ClientSecret, pcfClient.Config.SkipSslValidation)
	return eventsource.NewFirehose(ac, config), nil
}

// Nozzle creates a Nozzle object which glues the event source and event router
func (s *SplunkFirehoseNozzle) Nozzle(eventSource eventsource.Source, eventRouter eventrouter.Router) *nozzle.Nozzle {
	firehoseConfig := &nozzle.Config{
		Logger: s.logger,
	}

	return nozzle.New(eventSource, eventRouter, firehoseConfig)
}

// Run creates all necessary objects, reading events from PCF firehose and sending to target Splunk index
// It runs forever until something goes wrong
func (s *SplunkFirehoseNozzle) Run(shutdownChan chan os.Signal) error {
	eventSink, err := s.EventSink()
	if err != nil {
		s.logger.Error("Failed to create event sink", nil)
		return err
	}

	s.logger.Info("Running splunk-firehose-nozzle with following configuration variables ", s.config.ToMap())

	pcfClient, err := s.PCFClient()
	if err != nil {
		s.logger.Error("Failed to get info from PCF Server", nil)
		return err
	}

	appCache, err := s.AppCache(pcfClient)
	if err != nil {
		s.logger.Error("Failed to start App Cache", nil)
		return err
	}

	err = appCache.Open()
	if err != nil {
		s.logger.Error("Failed to open App Cache", nil)
		return err
	}
	defer appCache.Close()

	eventRouter, err := s.EventRouter(appCache, eventSink)
	if err != nil {
		s.logger.Error("Failed to create event router", nil)
		return err
	}

	eventSource, err := s.EventSource(pcfClient)

	if err != nil {
		return err
	}

	noz := s.Nozzle(eventSource, eventRouter)

	go func() {

	}()

	// Continuous Loop will run forever
	go func() {
		err := noz.Start()
		if err != nil {
			s.logger.Error("Firehose consumer exits with error", err)
		}
		shutdownChan <- os.Interrupt
	}()
	select {
	case <-shutdownChan:
		s.logger.Info("Splunk Nozzle is going to exit gracefully")
		noz.Close()
		return eventSink.Close()
	case gatewayError := <-gatewayErrChan:
		s.logger.Error("Error from PCF RLP gateway", gatewayError)
		noz.Close()
		eventSink.Close()
		return gatewayError
	}

}
