package main

import (
	"errors"
	"flag"
	"fmt"
	"log"

	"code.cloudfoundry.org/cflager"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/go-cfclient"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/drain"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/splunk"
)

var (
	debug = kingpin.
		Flag("debug", "Enable debug mode: forward to standard out intead of splunk").
		OverrideDefaultFromEnvar("DEBUG").Default("false").Bool()
	apiEndpoint = kingpin.
			Flag("api-endpoint", "Api endpoint address").
			OverrideDefaultFromEnvar("API_ENDPOINT").Required().String()
	user = kingpin.
		Flag("user", "Admin user.").
		OverrideDefaultFromEnvar("FIREHOSE_USER").Required().String()
	password = kingpin.
			Flag("password", "Admin password.").
			OverrideDefaultFromEnvar("FIREHOSE_PASSWORD").Required().String()
	dopplerEndpoint = kingpin.
			Flag("doppler-endpoint", "doppler endpoint, logging_endpoint in /v2/info").
			OverrideDefaultFromEnvar("DOPPLER_ENDPOINT").Required().String()
	skipSSLValidation = kingpin.
				Flag("skip-ssl-validation", "Please don't").
				OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Default("false").Bool()
	wantedEvents = kingpin.
			Flag("events", fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", eventRouting.GetListAuthorizedEventEvents())).
			OverrideDefaultFromEnvar("EVENTS").Default("ValueMetric,CounterEvent,ContainerMetric").String()
	keepAlive = kingpin.
			Flag("fh-keep-alive", "Keep Alive duration for the firehose consumer").
			OverrideDefaultFromEnvar("FH_KEEP_ALIVE").Default("25s").Duration()
	subscriptionId = kingpin.
			Flag("subscription-id", "Id for the subscription.").
			OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").Default("splunk-firehose").String()
	splunkToken = kingpin.
			Flag("splunk-token", "Splunk HTTP event collector token").
			OverrideDefaultFromEnvar("SPLUNK_TOKEN").Required().String()
	splunkHost = kingpin.
			Flag("splunk-host", "Splunk HTTP event collector host").
			OverrideDefaultFromEnvar("SPLUNK_HOST").Required().String()
	flushInterval = kingpin.
			Flag("flush-interval", "Every interval flushes to heavy forwarder every ").
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
		splunkCLient := splunk.NewSplunkClient(*splunkToken, *splunkHost, *skipSSLValidation, logger)
		loggingClient = drain.NewLoggingSplunk(logger, splunkCLient, *flushInterval)
	}

	events := eventRouting.NewEventRouting(caching.NewCachingEmpty() /*todo*/, loggingClient)
	err := events.SetupEventRouting(*wantedEvents)
	if err != nil {
		log.Fatal("Error setting up event routing: ", err)
	}

	cfConfig := &cfclient.Config{
		ApiAddress:        *apiEndpoint,
		Username:          *user,
		Password:          *password,
		SkipSslValidation: *skipSSLValidation,
	}
	cfClient := cfclient.NewClient(cfConfig)

	firehoseConfig := &splunk.FirehoseConfig{
		TrafficControllerURL:   *dopplerEndpoint,
		InsecureSSLSkipVerify:  *skipSSLValidation,
		IdleTimeoutSeconds:     *keepAlive,
		FirehoseSubscriptionID: *subscriptionId,
	}

	//todo: replace firehose-to-syslog client, get token via uaa
	if loggingClient.Connect() {
		firehoseClient := splunk.NewFirehoseNozzle(cfClient, events, firehoseConfig)
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
