package main

import (
	"flag"

	"code.cloudfoundry.org/cflager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"
)

var (
	version string
	branch  string
	commit  string
	buildos string
)

func main() {
	cflager.AddFlags(flag.CommandLine)

	logger, _ := cflager.New("splunk-nozzle-logger")
	logger.Info("Running splunk-firehose-nozzle")

	config, err := splunknozzle.NewConfigFromCmdFlags(version, branch, commit, buildos)
	if err != nil {
		logger.Error("Failed to get configuration", err)
		return
	}

	splunkNozzle := splunknozzle.NewSplunkFirehoseNozzle(config)
	err = splunkNozzle.Run(logger)
	if err != nil {
		logger.Error("Failed to run splunk-firehose-nozzle", err)
	}
}
