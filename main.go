package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/pivotal-golang/lager"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/auth"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/config"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/nozzle"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/splunk"
)

const flushWindow = time.Second * 10

func main() {
	cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := cf_lager.New("splunk-logger")
	logger.Info("Running splunk-firehose-nozzle")

	config, err := config.Parse()
	if err != nil {
		logger.Fatal("Unable to parse config", err)
	}

	token := getToken(config, logger)

	consumer := consumer.New(config.TrafficControllerURL, &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}, nil)
	events, errors := consumer.Firehose(config.FirehoseSubscriptionID, token)

	splunkEventSerializer := &splunk.SplunkEventSerializer{}
	splunkClient := splunk.NewSplunkClient(
		config.SplunkToken, config.SplunkHost, config.InsecureSkipVerify, logger,
	)

	logger.Info(fmt.Sprintf("Forwarding events: %s", config.SelectedEvents))
	forwarder := nozzle.NewForwarder(
		splunkClient, splunkEventSerializer,
		config.SelectedEvents, events, errors, logger,
	)
	err = forwarder.Run(flushWindow)
	if err != nil {
		logger.Fatal("Error forwarding", err)
	}
}

func getToken(config *config.Config, logger lager.Logger) string {
	fetcher := auth.NewUAATokenFetcher(config.UAAURL, config.Username, config.Password, true)
	token, err := fetcher.FetchAuthToken()
	if err != nil {
		logger.Fatal("Unable to fetch token", err)
	}

	return token
}
