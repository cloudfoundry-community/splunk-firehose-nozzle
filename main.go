package main

import (
	"crypto/tls"
	"flag"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cloudfoundry/noaa/consumer"
	"github.com/pivotal-golang/lager"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/auth"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/config"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/nozzle"
)

func main() {
	cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := cf_lager.New("logsearch-broker")
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
	//todo: do something more useful with error channel

	splunkClient := nozzle.NewSplunkClient(
		config.SplunkToken, config.SplunkHost, config.InsecureSkipVerify, logger,
	)
	forwarder := nozzle.NewSplunkForwarder(splunkClient, events, errors)
	err = forwarder.Run()
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
