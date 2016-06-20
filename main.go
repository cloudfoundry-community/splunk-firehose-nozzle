package main

import (
	"flag"

	"github.com/cloudfoundry-incubator/cf-lager"
	"github.com/cf-platform-eng/splunk-firehose-nozzle/config"
)

func main() {
	cf_lager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := cf_lager.New("logsearch-broker")
	logger.Info("Running splunk-firehose-nozzle")

	_, err := config.Parse()
	if err != nil {
		logger.Fatal("Unable to parse config", err)
	}
}
