package main

import (
	"flag"
	"fmt"

	"code.cloudfoundry.org/cflager"
	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"
	"gopkg.in/alecthomas/kingpin.v2"
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
		OverrideDefaultFromEnvar("JOB_HOST").Default("").String()

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
	hecWorkers = kingpin.Flag("hec-workers", "How many workers (concurrency) when post data to HEC").
			OverrideDefaultFromEnvar("HEC_WORKERS").Default("8").Int()
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

	splunkNozzle := splunknozzle.NewSplunkFirehoseNozzle(*apiEndpoint, *user, *password, *splunkHost, *splunkToken, *splunkIndex)

	// Setup all other params
	splunkNozzle.WithJobName(*jobName).
		WithJobIndex(*jobIndex).
		WithJobHost(*jobHost).
		WithSkipSSL(*skipSSL).
		WithSubscriptionID(*subscriptionId).
		WithKeepAlive(*keepAlive).
		WithAddAppInfo(*addAppInfo).
		WithBoltDBPath(*boltDBPath).
		WithWantedEvents(*wantedEvents).
		WithExtraFields(*extraFields).
		WithFlushInterval(*flushInterval).
		WithQueueSize(*queueSize).
		WithBatchSize(*batchSize).
		WithRetries(*retries).
		WithHecWorkers(*hecWorkers).
		WithVersion(version).
		WithBranch(branch).
		WithCommit(commit).
		WithBuildOS(buildos).
		WithDebug(*debug)

	err := splunkNozzle.Run(logger)
	if err != nil {
		logger.Error("Failed to run splunk-firehose-nozzle", err)
	}
}
