package main

import (
	"flag"

	"code.cloudfoundry.org/cflager"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"
	"gopkg.in/alecthomas/kingpin.v2"
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

	kingpin.Version(version)
	kingpin.Parse()

	logger, _ := cflager.New("splunk-nozzle-logger")
	logger.Info("Running splunk-firehose-nozzle")

	config := splunknozzle.NewConfigFromCmdFlags()
	config.Version = version
	config.Branch = branch
	config.Commit = commit
	config.BuildOS = buildos

	splunkNozzle := splunknozzle.NewSplunkFirehoseNozzle(config)
	err := splunkNozzle.Run(logger)
	if err != nil {
		logger.Error("Failed to run splunk-firehose-nozzle", err)
	}
}
