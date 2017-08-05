package splunknozzle

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/events"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type Config struct {
	ApiEndpoint string
	User        string
	Password    string

	SplunkToken string
	SplunkHost  string
	SplunkIndex string

	JobName  string
	JobIndex string
	JobHost  string

	SkipSSL        bool
	SubscriptionID string
	KeepAlive      time.Duration

	AddAppInfo         bool
	IgnoreMissingApps  bool
	MissingAppCacheTTL time.Duration
	AppCacheTTL        time.Duration

	BoltDBPath   string
	WantedEvents string
	ExtraFields  string

	FlushInterval time.Duration
	QueueSize     int
	BatchSize     int
	Retries       int
	HecWorkers    int

	Version string
	Branch  string
	Commit  string
	BuildOS string

	Debug bool
}

func NewConfigFromCmdFlags(version, branch, commit, buildos string) *Config {
	c := &Config{}

	c.Version = version
	c.Branch = branch
	c.Commit = commit
	c.BuildOS = buildos

	kingpin.Version(version)
	kingpin.Flag("debug", "Enable debug mode: forward to standard out intead of splunk").
		OverrideDefaultFromEnvar("DEBUG").Default("false").BoolVar(&c.Debug)
	kingpin.Flag("skip-ssl-validation", "Skip cert validation (for dev environments").
		OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Default("false").BoolVar(&c.SkipSSL)
	kingpin.Flag("job-name", "Job name to tag nozzle's own log events").
		OverrideDefaultFromEnvar("JOB_NAME").Default("splunk-nozzle").StringVar(&c.JobName)
	kingpin.Flag("job-index", "Job index to tag nozzle's own log events").
		OverrideDefaultFromEnvar("JOB_INDEX").Default("-1").StringVar(&c.JobIndex)
	kingpin.Flag("job-host", "Job host to tag nozzle's own log events").
		OverrideDefaultFromEnvar("JOB_HOST").Default("").StringVar(&c.JobHost)

	kingpin.Flag("add-app-info", "Query API to fetch app details").
		OverrideDefaultFromEnvar("ADD_APP_INFO").Default("false").BoolVar(&c.AddAppInfo)
	kingpin.Flag("ignore-missing-app", "If app is missing, if stop repeatly querying app info from PCF").
		OverrideDefaultFromEnvar("IGNORE-MISSING-APP").Default("true").BoolVar(&c.IgnoreMissingApps)
	kingpin.Flag("missing-app-cache-invalidate-ttl", "How frequently the missing app info cache invalidates").
		OverrideDefaultFromEnvar("MISSING_APP_CACHE_INVALIDATE_TTL").Default("0s").DurationVar(&c.MissingAppCacheTTL)

	kingpin.Flag("app-cache-invalidate-ttl", "How frequently the app info local cache invalidates").
		OverrideDefaultFromEnvar("APP_CACHE_INVALIDATE_TTL").Default("0s").DurationVar(&c.AppCacheTTL)

	kingpin.Flag("api-endpoint", "API endpoint address").
		OverrideDefaultFromEnvar("API_ENDPOINT").Required().StringVar(&c.ApiEndpoint)
	kingpin.Flag("user", "Admin user.").
		OverrideDefaultFromEnvar("API_USER").Required().StringVar(&c.User)
	kingpin.Flag("password", "Admin password.").
		OverrideDefaultFromEnvar("API_PASSWORD").Required().StringVar(&c.Password)
	kingpin.Flag("boltdb-path", "Bolt Database path ").
		Default("cache.db").OverrideDefaultFromEnvar("BOLTDB_PATH").StringVar(&c.BoltDBPath)

	kingpin.Flag("events", fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", events.AuthorizedEvents())).OverrideDefaultFromEnvar("EVENTS").Default("ValueMetric,CounterEvent,ContainerMetric").StringVar(&c.WantedEvents)
	kingpin.Flag("extra-fields", "Extra fields you want to annotate your events with, example: '--extra-fields=env:dev,something:other ").
		OverrideDefaultFromEnvar("EXTRA_FIELDS").Default("").StringVar(&c.ExtraFields)
	kingpin.Flag("firehose-keep-alive", "Keep Alive duration for the firehose consumer").
		OverrideDefaultFromEnvar("FIREHOSE_KEEP_ALIVE").Default("25s").DurationVar(&c.KeepAlive)
	kingpin.Flag("subscription-id", "Id for the subscription.").
		OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").Default("splunk-firehose").StringVar(&c.SubscriptionID)

	kingpin.Flag("splunk-token", "Splunk HTTP event collector token").
		OverrideDefaultFromEnvar("SPLUNK_TOKEN").Required().StringVar(&c.SplunkToken)
	kingpin.Flag("splunk-host", "Splunk HTTP event collector host").
		OverrideDefaultFromEnvar("SPLUNK_HOST").Required().StringVar(&c.SplunkHost)
	kingpin.Flag("splunk-index", "Splunk index").
		OverrideDefaultFromEnvar("SPLUNK_INDEX").Required().StringVar(&c.SplunkIndex)
	kingpin.Flag("flush-interval", "Every interval flushes to heavy forwarder every").
		OverrideDefaultFromEnvar("FLUSH_INTERVAL").Default("5s").DurationVar(&c.FlushInterval)
	kingpin.Flag("consumer-queue-size", "Consumer queue buffer size").
		OverrideDefaultFromEnvar("CONSUMER_QUEUE_SIZE").Default("10000").IntVar(&c.QueueSize)
	kingpin.Flag("hec-batch-size", "Batchsize of the events pushing to HEC").
		OverrideDefaultFromEnvar("HEC_BATCH_SIZE").Default("1000").IntVar(&c.BatchSize)
	kingpin.Flag("hec-retries", "Number of retries before dropping events").
		OverrideDefaultFromEnvar("HEC_RETRIES").Default("5").IntVar(&c.Retries)
	kingpin.Flag("hec-workers", "How many workers (concurrency) when post data to HEC").
		OverrideDefaultFromEnvar("HEC_WORKERS").Default("8").IntVar(&c.HecWorkers)

	kingpin.Parse()

	return c
}
