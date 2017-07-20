package splunknozzle

import (
	"crypto/tls"
	"fmt"
	"time"

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
	apiEndpoint string
	user        string
	password    string

	splunkToken string
	splunkHost  string
	splunkIndex string

	jobName  string
	jobIndex string
	jobHost  string

	skipSSL        bool
	subscriptionID string
	keepAlive      time.Duration

	addAppInfo   bool
	boltDBPath   string
	wantedEvents string
	extraFields  string

	flushInterval time.Duration
	queueSize     int
	batchSize     int
	retries       int
	hecWorkers    int

	version string
	branch  string
	commit  string
	buildos string

	debug bool
}

func NewSplunkFirehoseNozzle(apiEndpoint, user, password, splunkHost, splunkToken, splunkIndex string) *SplunkFirehoseNozzle {
	return &SplunkFirehoseNozzle{
		apiEndpoint: apiEndpoint,
		user:        user,
		password:    password,

		splunkToken: splunkToken,
		splunkHost:  splunkHost,
		splunkIndex: splunkIndex,

		jobName:  "splunk-nozzle",
		jobIndex: "-1",
		jobHost:  "",

		skipSSL:        false,
		subscriptionID: "splunk-firehose",
		keepAlive:      time.Second * 25,

		addAppInfo:   false,
		boltDBPath:   "cache.db",
		wantedEvents: "ValueMetric,CounterEvent,ContainerMetric",
		extraFields:  "",

		flushInterval: time.Second * 5,
		queueSize:     10000,
		batchSize:     1000,
		retries:       5,
		hecWorkers:    8,

		debug: false,
	}
}

func (s *SplunkFirehoseNozzle) ApiEndpoint() string {
	return s.apiEndpoint
}

func (s *SplunkFirehoseNozzle) User() string {
	return s.user
}

func (s *SplunkFirehoseNozzle) Password() string {
	return s.password
}

func (s *SplunkFirehoseNozzle) SplunkHost() string {
	return s.splunkHost
}

func (s *SplunkFirehoseNozzle) SplunkToken() string {
	return s.splunkToken
}

func (s *SplunkFirehoseNozzle) SplunkIndex() string {
	return s.splunkIndex
}

func (s *SplunkFirehoseNozzle) WithJobName(jobName string) *SplunkFirehoseNozzle {
	s.jobName = jobName
	return s
}

func (s *SplunkFirehoseNozzle) JobName() string {
	return s.jobName
}

func (s *SplunkFirehoseNozzle) WithJobIndex(jobIndex string) *SplunkFirehoseNozzle {
	s.jobIndex = jobIndex
	return s
}

func (s *SplunkFirehoseNozzle) JobIndex() string {
	return s.jobIndex
}

func (s *SplunkFirehoseNozzle) WithJobHost(jobHost string) *SplunkFirehoseNozzle {
	s.jobHost = jobHost
	return s
}

func (s *SplunkFirehoseNozzle) JobHost() string {
	return s.jobHost
}

func (s *SplunkFirehoseNozzle) WithSkipSSL(skipSSL bool) *SplunkFirehoseNozzle {
	s.skipSSL = skipSSL
	return s
}

func (s *SplunkFirehoseNozzle) SkipSSL() bool {
	return s.skipSSL
}

func (s *SplunkFirehoseNozzle) WithSubscriptionID(subID string) *SplunkFirehoseNozzle {
	s.subscriptionID = subID
	return s
}

func (s *SplunkFirehoseNozzle) SubscriptionID() string {
	return s.subscriptionID
}

func (s *SplunkFirehoseNozzle) WithKeepAlive(keepAlive time.Duration) *SplunkFirehoseNozzle {
	s.keepAlive = keepAlive
	return s
}

func (s *SplunkFirehoseNozzle) KeepAlive() time.Duration {
	return s.keepAlive
}

func (s *SplunkFirehoseNozzle) WithAddAppInfo(addAppInfo bool) *SplunkFirehoseNozzle {
	s.addAppInfo = addAppInfo
	return s
}

func (s *SplunkFirehoseNozzle) AddAppInfo() bool {
	return s.addAppInfo
}

func (s *SplunkFirehoseNozzle) WithBoltDBPath(path string) *SplunkFirehoseNozzle {
	s.boltDBPath = path
	return s
}

func (s *SplunkFirehoseNozzle) BoltDBPath() string {
	return s.boltDBPath
}

func (s *SplunkFirehoseNozzle) WithWantedEvents(events string) *SplunkFirehoseNozzle {
	s.wantedEvents = events
	return s
}

func (s *SplunkFirehoseNozzle) WantedEvents() string {
	return s.wantedEvents
}

func (s *SplunkFirehoseNozzle) WithExtraFields(fields string) *SplunkFirehoseNozzle {
	s.extraFields = fields
	return s
}

func (s *SplunkFirehoseNozzle) ExtraFields() string {
	return s.extraFields
}

func (s *SplunkFirehoseNozzle) WithFlushInterval(interval time.Duration) *SplunkFirehoseNozzle {
	s.flushInterval = interval
	return s
}

func (s *SplunkFirehoseNozzle) FlushInterval() time.Duration {
	return s.flushInterval
}

func (s *SplunkFirehoseNozzle) WithQueueSize(queueSize int) *SplunkFirehoseNozzle {
	s.queueSize = queueSize
	return s
}

func (s *SplunkFirehoseNozzle) QueueSize() int {
	return s.queueSize
}

func (s *SplunkFirehoseNozzle) WithBatchSize(batchSize int) *SplunkFirehoseNozzle {
	s.batchSize = batchSize
	return s
}

func (s *SplunkFirehoseNozzle) BatchSize() int {
	return s.batchSize
}

func (s *SplunkFirehoseNozzle) WithHecWorkers(hecWorkers int) *SplunkFirehoseNozzle {
	s.hecWorkers = hecWorkers
	return s
}

func (s *SplunkFirehoseNozzle) HecWorkers() int {
	return s.hecWorkers
}

func (s *SplunkFirehoseNozzle) WithRetries(retries int) *SplunkFirehoseNozzle {
	s.retries = retries
	return s
}

func (s *SplunkFirehoseNozzle) Retries() int {
	return s.retries
}

func (s *SplunkFirehoseNozzle) WithVersion(version string) *SplunkFirehoseNozzle {
	s.version = version
	return s
}

func (s *SplunkFirehoseNozzle) Version() string {
	return s.version
}

func (s *SplunkFirehoseNozzle) WithBranch(branch string) *SplunkFirehoseNozzle {
	s.branch = branch
	return s
}

func (s *SplunkFirehoseNozzle) Branch() string {
	return s.branch
}

func (s *SplunkFirehoseNozzle) WithCommit(commit string) *SplunkFirehoseNozzle {
	s.commit = commit
	return s
}

func (s *SplunkFirehoseNozzle) Commit() string {
	return s.commit
}

func (s *SplunkFirehoseNozzle) WithBuildOS(buildos string) *SplunkFirehoseNozzle {
	s.buildos = buildos
	return s
}

func (s *SplunkFirehoseNozzle) BuildOS() string {
	return s.buildos
}

func (s *SplunkFirehoseNozzle) WithDebug(debug bool) *SplunkFirehoseNozzle {
	s.debug = debug
	return s
}

func (s *SplunkFirehoseNozzle) Debug() bool {
	return s.debug
}

// eventRouting creates eventRouting object and setup routings for interested events
func (s *SplunkFirehoseNozzle) eventRouting(cache caching.Caching, logClient logging.Logging) (*eventRouting.EventRouting, error) {
	events := eventRouting.NewEventRouting(cache, logClient)
	err := events.SetupEventRouting(s.wantedEvents)
	if err != nil {
		return nil, err
	}
	return events, nil
}

// pcfClient creates a client object which can talk to PCF
func (s *SplunkFirehoseNozzle) pcfClient() (*cfclient.Client, error) {
	cfConfig := &cfclient.Config{
		ApiAddress:        s.apiEndpoint,
		Username:          s.user,
		Password:          s.password,
		SkipSslValidation: s.skipSSL,
	}

	cfClient, err := cfclient.NewClient(cfConfig)
	if err != nil {
		return nil, err
	}
	return cfClient, nil
}

// appCache creates inmemory cache or boltDB cache
func (s *SplunkFirehoseNozzle) appCache(cfClient *cfclient.Client) caching.Caching {
	if s.addAppInfo {
		cache := caching.NewCachingBolt(cfClient, s.boltDBPath)
		cache.CreateBucket()
		return cache
	}

	return caching.NewCachingEmpty()
}

// logClient creates std logging or Splunk logging
func (s *SplunkFirehoseNozzle) logClient(logger lager.Logger) (logging.Logging, error) {
	if s.debug {
		return &drain.LoggingStd{}, nil
	}

	parsedExtraFields, err := extrafields.ParseExtraFields(s.extraFields)
	if err != nil {
		return nil, err
	}

	// SplunkClient for nozzle internal logging
	splunkClient := splunk.NewSplunkClient(s.splunkToken, s.splunkHost, s.splunkIndex, parsedExtraFields, s.skipSSL, logger)
	logger.RegisterSink(sink.NewSplunkSink(s.jobName, s.jobIndex, s.jobHost, splunkClient))

	// SplunkClients for raw event POST
	var splunkClients []splunk.SplunkClient
	for i := 0; i < s.hecWorkers; i++ {
		splunkClient := splunk.NewSplunkClient(s.splunkToken, s.splunkHost, s.splunkIndex, parsedExtraFields, s.skipSSL, logger)
		splunkClients = append(splunkClients, splunkClient)
	}

	loggingConfig := &drain.LoggingConfig{
		FlushInterval: s.flushInterval,
		QueueSize:     s.queueSize,
		BatchSize:     s.batchSize,
		Retries:       s.retries,
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
	consumer := consumer.New(dopplerEndpoint, &tls.Config{InsecureSkipVerify: s.skipSSL}, nil)
	consumer.RefreshTokenFrom(tokenRefresher)
	consumer.SetIdleTimeout(s.keepAlive)
	return consumer
}

// firehoseClient creates FirehoseNozzle object which glues the event source and event sink
func (s *SplunkFirehoseNozzle) firehoseClient(consumer *consumer.Consumer, events *eventRouting.EventRouting) *firehoseclient.FirehoseNozzle {
	firehoseConfig := &firehoseclient.FirehoseConfig{
		FirehoseSubscriptionID: s.subscriptionID,
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
		"version":        s.version,
		"branch":         s.branch,
		"commit":         s.commit,
		"buildos":        s.buildos,
		"flush-interval": s.flushInterval,
		"queue-size":     s.queueSize,
		"batch-size":     s.batchSize,
		"workers":        s.hecWorkers,
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
