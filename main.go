package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"code.cloudfoundry.org/cflager"
	"code.cloudfoundry.org/lager"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/go-cfclient"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/auth"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/drain"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/firehoseclient"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/splunk"
)

var (
	uaaEndpoint = kingpin.Flag("uaa-endpoint", "UAA endpoint address").
			OverrideDefaultFromEnvar("UAA_ENDPOINT").Required().String()
	uaaUser = kingpin.Flag("uaa-user", "Admin user.").
		OverrideDefaultFromEnvar("FIREHOSE_UAA_USER").Required().String()
	uaaSecret = kingpin.Flag("uaa-secret", "Admin password.").
			OverrideDefaultFromEnvar("FIREHOSE_UAA_SECRET").Required().String()
	skipSSL = kingpin.Flag("skip-ssl-validation", "Please don't").
		OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Default("false").Bool()

	addAppInfo = kingpin.Flag("add-app-info", "Query API to fetch app details").
			OverrideDefaultFromEnvar("ADD_APP_INFO").Default("false").Bool()
	apiEndpoint = kingpin.Flag("api-endpoint", "API endpoint address").
			OverrideDefaultFromEnvar("API_ENDPOINT").String()
	user = kingpin.Flag("user", "Admin user.").
		OverrideDefaultFromEnvar("API_USER").String()
	password = kingpin.Flag("password", "Admin password.").
			OverrideDefaultFromEnvar("API_PASSWORD").String()
	boltDBPath = kingpin.Flag("boltdb-path", "Bolt Database path ").
			Default("cache.db").OverrideDefaultFromEnvar("BOLTDB_PATH").String()

	debug = kingpin.Flag("debug", "Enable debug mode: forward to standard out intead of splunk").
		OverrideDefaultFromEnvar("DEBUG").Default("false").Bool()
	dopplerEndpoint = kingpin.Flag("doppler-endpoint", "doppler endpoint, logging_endpoint in /v2/info").
			OverrideDefaultFromEnvar("DOPPLER_ENDPOINT").Required().String()
	wantedEvents = kingpin.Flag("events", fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", eventRouting.GetListAuthorizedEventEvents())).
			OverrideDefaultFromEnvar("EVENTS").Default("ValueMetric,CounterEvent,ContainerMetric").String()
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
)

var (
	version = "0.0.1"
)

func main() {
	cflager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := cflager.New("splunk-logger")
	logger.Info("Running splunk-firehose-nozzle")

	kingpin.Version(version)
	kingpin.Parse()

	var loggingClient logging.Logging
	if *debug {
		loggingClient = &drain.LoggingStd{}
	} else {
		splunkCLient := splunk.NewSplunkClient(*splunkToken, *splunkHost, *splunkIndex, *skipSSL, logger)
		loggingClient = drain.NewLoggingSplunk(logger, splunkCLient, *flushInterval)
	}

	logger.Info("Setting up caching")
	cache := setupCache(logger)

	logger.Info("Setting up event routing")
	events := eventRouting.NewEventRouting(cache, loggingClient)
	err := events.SetupEventRouting(*wantedEvents)
	if err != nil {
		log.Fatal("Error setting up event routing: ", err)
	}

	tokenRefresher := auth.NewUAATokenFetcher(*uaaEndpoint, *uaaUser, *uaaSecret, *skipSSL)
	firehoseConfig := &firehoseclient.FirehoseConfig{
		TrafficControllerURL:   *dopplerEndpoint,
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

func setupCache(logger lager.Logger) caching.Caching {
	if *addAppInfo {
		if *user == "" {
			logger.Fatal("Unable to setup cache", errors.New("User is required when add-app-info is true"))
		}
		if *password == "" {
			logger.Fatal("Unable to setup cache", errors.New("Password is required when add-app-info is true"))
		}
		if *apiEndpoint == "" {
			logger.Fatal("Unable to setup cache", errors.New("ApiEndpoint is required when add-app-info is true"))
		}

		logger.Info("Connecting to Cloud Foundry")
		cfConfig := &cfclient.Config{
			ApiAddress:        *apiEndpoint,
			Username:          *user,
			Password:          *password,
			SkipSslValidation: *skipSSL,
		}
		cfClient := cfclient.NewClient(cfConfig)

		cache := caching.NewCachingBolt(cfClient, *boltDBPath)
		cache.CreateBucket()

		return cache
	} else {
		return caching.NewCachingEmpty()
	}
}
