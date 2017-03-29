package main

import (
	"errors"
	"flag"
	"log"

	"code.cloudfoundry.org/cflager"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/extrafields"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/go-cfclient"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/auth"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/config"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/drain"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/firehoseclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/sink"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunk"
)

func main() {
	cflager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := cflager.New("splunk-nozzle-logger")
	logger.Info("Running splunk-firehose-nozzle")

	c, err := config.Parse()

	parsedExtraFields, err := extrafields.ParseExtraFields(c.ExtraFields)
	if err != nil {
		log.Fatal("Error parsing etra fields: ", err)
	}

	var loggingClient logging.Logging
	if c.Debug {
		loggingClient = &drain.LoggingStd{}
	} else {
		splunkCLient := splunk.NewSplunkClient(c.SplunkToken, c.SplunkHost, parsedExtraFields, c.SkipSSL, logger)
		loggingClient = drain.NewLoggingSplunk(logger, splunkCLient, c.FlushInterval, c.SplunkIndex, c.MappingList.Mappings)
		logger.RegisterSink(sink.NewSplunkSink(c.JobName, c.JobIndex, c.JobHost, splunkCLient))
	}

	logger.Info("Connecting to Cloud Foundry")
	cfConfig := &cfclient.Config{
		ApiAddress:        c.ApiEndpoint,
		Username:          c.ApiUser,
		Password:          c.ApiPassword,
		SkipSslValidation: c.SkipSSL,
	}
	cfClient, err := cfclient.NewClient(cfConfig)
	if err != nil {
		log.Fatal("Error setting up cf client: ", err)
	}

	logger.Info("Setting up caching")
	var cache caching.Caching
	if c.AddAppInfo {
		cache = caching.NewCachingBolt(cfClient, c.BoldDBPath)
		cache.CreateBucket()
	} else {
		cache = caching.NewCachingEmpty()
	}

	logger.Info("Setting up event routing")
	events := eventRouting.NewEventRouting(cache, loggingClient)
	err = events.SetupEventRouting(c.WantedEvents)
	if err != nil {
		log.Fatal("Error setting up event routing: ", err)
	}

	tokenRefresher := auth.NewTokenRefreshAdapter(cfClient)
	dopplerEndpoint := cfClient.Endpoint.DopplerEndpoint
	firehoseConfig := &firehoseclient.FirehoseConfig{
		TrafficControllerURL:   dopplerEndpoint,
		InsecureSSLSkipVerify:  c.SkipSSL,
		IdleTimeoutSeconds:     c.KeepAlive,
		FirehoseSubscriptionID: c.SubscriptionId,
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
