package main

import (
	"errors"
	"flag"
	"fmt"

	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/extrafields"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/go-cfclient"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/auth"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/drain"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/firehoseclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/sink"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
)

var (
	debug = kingpin.Flag("debug", "Enable debug mode: forward to standard out intead of splunk").
		OverrideDefaultFromEnvar("DEBUG").Default("false").Bool()
	skipSSL = kingpin.Flag("skip-ssl-validation", "Skip cert validation (for dev environments").
		OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Default("false").Bool()
	jobName = kingpin.Flag("job-name", "Job name to tag nozzle's own log events").
		OverrideDefaultFromEnvar("JOB_NAME").Default("splunk-nozzle").String()
	jobIndex = kingpin.Flag("job-index", "Job index to tag nozzle's own log events").
			OverrideDefaultFromEnvar("JOB_INDEX").Default("-1").String()
	jobHost = kingpin.Flag("job-host", "Job host to tag nozzle's own log events").
		OverrideDefaultFromEnvar("JOB_HOST").Default("localhost").String()

	addAppInfo = kingpin.Flag("add-app-info", "Query API to fetch app details").
			OverrideDefaultFromEnvar("ADD_APP_INFO").Default("false").Bool()
	apiEndpoint = kingpin.Flag("api-endpoint", "API endpoint address").
			OverrideDefaultFromEnvar("API_ENDPOINT").Required().String()
	user = kingpin.Flag("user", "Admin user.").
		OverrideDefaultFromEnvar("API_USER").Required().String()
	password = kingpin.Flag("password", "Admin password.").
			OverrideDefaultFromEnvar("API_PASSWORD").Required().String()
	boltDBPath = kingpin.Flag("boltdb-path", "Bolt Database path ").
			Default("cache.db").OverrideDefaultFromEnvar("BOLTDB_PATH").String()

	wantedEvents = kingpin.Flag("events", fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", eventRouting.GetListAuthorizedEventEvents())).
			OverrideDefaultFromEnvar("EVENTS").Default("ValueMetric,CounterEvent,ContainerMetric").String()
	extraFields = kingpin.Flag("extra-fields", "Extra fields you want to annotate your events with, example: '--extra-fields=env:dev,something:other ").
			OverrideDefaultFromEnvar("EXTRA_FIELDS").Default("").String()
	keepAlive = kingpin.Flag("firehose-keep-alive", "Keep Alive duration for the firehose consumer").
			OverrideDefaultFromEnvar("FIREHOSE_KEEP_ALIVE").Default("25s").Duration()
	subscriptionId = kingpin.Flag("subscription-id", "Id for the subscription.").
			OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").Default("splunk-firehose").String()

	splunkToken = kingpin.Flag("splunk-token", "Splunk HTTP event collector token").
			OverrideDefaultFromEnvar("SPLUNK_TOKEN").Required().String()
	splunkHost = kingpin.Flag("splunk-host", "Splunk HTTP event collector host").
			OverrideDefaultFromEnvar("SPLUNK_HOST").Required().String()
	splunkIndex = kingpin.Flag("splunk-index", "Splunk index").
			OverrideDefaultFromEnvar("SPLUNK_INDEX").Required().String()
	flushInterval = kingpin.Flag("flush-interval", "Every interval flushes to heavy forwarder every ").
			OverrideDefaultFromEnvar("FLUSH_INTERVAL").Default("5s").Duration()
	queueSize = kingpin.Flag("consumer-queue-size", "Consumer queue buffer size").
			OverrideDefaultFromEnvar("CONSUMER_QUEUE_SIZE").Default("10000").Int()
	batchSize = kingpin.Flag("hec-batch-size", "Batchsize of the events pushing to HEC  ").
			OverrideDefaultFromEnvar("HEC_BATCH_SIZE").Default("1000").Int()
	retries = kingpin.Flag("hec-retries", "Number of retries before dropping events").
		OverrideDefaultFromEnvar("HEC_RETRIES").Default("5").Int()
)

var (
	version string
	branch  string
	commit  string
	buildos string
)

func main() {
	cflager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := cflager.New("splunk-nozzle-logger")
	logger.Info("Running splunk-firehose-nozzle")

	kingpin.Version(version)
	kingpin.Parse()

	parsedExtraFields, err := extrafields.ParseExtraFields(*extraFields)
	if err != nil {
		logger.Fatal("Error parsing extra fields: ", err)
	}

	loggingConfig := &drain.LoggingConfig{
		FlushInterval: *flushInterval,
		QueueSize:     *queueSize,
		BatchSize:     *batchSize,
		Retries:       *retries,
	}

	var loggingClient logging.Logging
	if *debug {
		loggingClient = &drain.LoggingStd{}
	} else {
		splunkCLient := splunk.NewSplunkClient(*splunkToken, *splunkHost, *splunkIndex, parsedExtraFields, *skipSSL, logger)
		loggingClient = drain.NewLoggingSplunk(logger, splunkCLient, loggingConfig)
		logger.RegisterSink(sink.NewSplunkSink(*jobName, *jobIndex, *jobHost, splunkCLient))
	}

	versionInfo := lager.Data{
		"version": version,
		"branch":  branch,
		"commit":  commit,
		"buildos": buildos,
	}

	logger.Info("Connecting to Cloud Foundry. splunk-firehose-nozzle runs", versionInfo)
	cfConfig := &cfclient.Config{
		ApiAddress:        *apiEndpoint,
		Username:          *user,
		Password:          *password,
		SkipSslValidation: *skipSSL,
	}
	cfClient, err := cfclient.NewClient(cfConfig)
	if err != nil {
		logger.Fatal("Error setting up cf client: ", err)
	}

	logger.Info("Setting up caching")
	var cache caching.Caching
	if *addAppInfo {
		cache = caching.NewCachingBolt(cfClient, *boltDBPath)
		cache.CreateBucket()
	} else {
		cache = caching.NewCachingEmpty()
	}

	logger.Info("Setting up event routing")
	events := eventRouting.NewEventRouting(cache, loggingClient)
	err = events.SetupEventRouting(*wantedEvents)
	if err != nil {
		logger.Fatal("Error setting up event routing: ", err)
	}

	tokenRefresher := auth.NewTokenRefreshAdapter(cfClient)
	dopplerEndpoint := cfClient.Endpoint.DopplerEndpoint
	firehoseConfig := &firehoseclient.FirehoseConfig{
		TrafficControllerURL:   dopplerEndpoint,
		InsecureSSLSkipVerify:  *skipSSL,
		IdleTimeoutSeconds:     *keepAlive,
		FirehoseSubscriptionID: *subscriptionId,
	}

	logger.Info("Connecting logging client")
	if loggingClient.Connect() {
		firehoseClient := firehoseclient.NewFirehoseNozzle(tokenRefresher, events, firehoseConfig)
		err := firehoseClient.Start()
		if err != nil {
			logger.Fatal("Failed connecting to Firehose", err)
		} else {
			logger.Info("Firehose Subscription Succesfull; routing events.")
		}
	} else {
		logger.Fatal("Failed connecting to Splunk", errors.New(""))
	}
}
