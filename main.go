package main

import (
	"flag"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/pivotal-golang/lager"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/auth"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/config"
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
	println(token)
}

func getToken(config *config.Config, logger lager.Logger) string {
	fetcher := auth.NewUAATokenFetcher(config.UAAURL, config.Username, config.Password, true)
	token, err := fetcher.FetchAuthToken()
	if err != nil {
		logger.Fatal("Unable to fetch token", err)
	}

	return token
}
