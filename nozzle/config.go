package splunknozzle

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/extrafields"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/drain"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type Config struct {
	ApiEndpoint string
	User        string
	Password    string

	SplunkToken  string
	SplunkHost   string
	IndexMapping *drain.IndexMapConfig

	JobName  string
	JobIndex string
	JobHost  string

	SkipSSL        bool
	SubscriptionID string
	KeepAlive      time.Duration

	AddAppInfo   bool
	BoltDBPath   string
	WantedEvents string
	ExtraFields  map[string]string

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

func NewConfigFromCmdFlags(version, branch, commit, buildos string) (*Config, error) {
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
	kingpin.Flag("firehose-keep-alive", "Keep Alive duration for the firehose consumer").
		OverrideDefaultFromEnvar("FIREHOSE_KEEP_ALIVE").Default("25s").DurationVar(&c.KeepAlive)
	kingpin.Flag("subscription-id", "Id for the subscription.").
		OverrideDefaultFromEnvar("FIREHOSE_SUBSCRIPTION_ID").Default("splunk-firehose").StringVar(&c.SubscriptionID)

	kingpin.Flag("job-name", "Job name to tag nozzle's own log events").
		OverrideDefaultFromEnvar("JOB_NAME").Default("splunk-nozzle").StringVar(&c.JobName)
	kingpin.Flag("job-index", "Job index to tag nozzle's own log events").
		OverrideDefaultFromEnvar("JOB_INDEX").Default("-1").StringVar(&c.JobIndex)
	kingpin.Flag("job-host", "Job host to tag nozzle's own log events").
		OverrideDefaultFromEnvar("JOB_HOST").Default("").StringVar(&c.JobHost)

	kingpin.Flag("api-endpoint", "API endpoint address").
		OverrideDefaultFromEnvar("API_ENDPOINT").Required().StringVar(&c.ApiEndpoint)
	kingpin.Flag("user", "Admin user.").
		OverrideDefaultFromEnvar("API_USER").Required().StringVar(&c.User)
	kingpin.Flag("password", "Admin password.").
		OverrideDefaultFromEnvar("API_PASSWORD").Required().StringVar(&c.Password)

	validEvents := fmt.Sprintf("Comma separated list of events you would like. Valid options are %s", eventRouting.GetListAuthorizedEventEvents())
	kingpin.Flag("events", validEvents).OverrideDefaultFromEnvar("EVENTS").Default("ValueMetric,CounterEvent,ContainerMetric").StringVar(&c.WantedEvents)

	kingpin.Flag("boltdb-path", "Bolt Database path ").
		Default("cache.db").OverrideDefaultFromEnvar("BOLTDB_PATH").StringVar(&c.BoltDBPath)
	kingpin.Flag("add-app-info", "Query API to fetch app details").
		OverrideDefaultFromEnvar("ADD_APP_INFO").Default("false").BoolVar(&c.AddAppInfo)
	var extraFields string
	kingpin.Flag("extra-fields", "Extra fields you want to annotate your events with, example: '--extra-fields=env:dev,something:other ").
		OverrideDefaultFromEnvar("EXTRA_FIELDS").Default("").StringVar(&extraFields)

	kingpin.Flag("splunk-token", "Splunk HTTP event collector token").
		OverrideDefaultFromEnvar("SPLUNK_TOKEN").Required().StringVar(&c.SplunkToken)
	kingpin.Flag("splunk-host", "Splunk HTTP event collector host").
		OverrideDefaultFromEnvar("SPLUNK_HOST").Required().StringVar(&c.SplunkHost)

	var indexMapping string
	kingpin.Flag("splunk-index-mapping", "Route the events to different Splunk indexes according to org, space, appid etc. Refer to doc for details.").
		OverrideDefaultFromEnvar("SPLUNK_INDEX_MAPPING").Required().StringVar(&indexMapping)

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

	err := c.ParseExtraFields(extraFields)
	if err != nil {
		return nil, err
	}

	err = c.ParseIndexMapping(indexMapping)
	if err != nil {
		return nil, err
	}

	if c.IndexMapping != nil {
		// We will need app info for org/space/app lookup
		c.AddAppInfo = true
	}
	return c, nil
}

func (c *Config) ParseExtraFields(extraFields string) error {
	parsedExtraFields, err := extrafields.ParseExtraFields(extraFields)
	if err != nil {
		return err
	}

	c.ExtraFields = parsedExtraFields
	return nil
}

func (c *Config) ParseIndexMapping(indexMapping string) error {
	if indexMapping == "" {
		return errors.New("Index mapping is required")
	}

	var mapConfig drain.IndexMapConfig
	err := json.Unmarshal([]byte(indexMapping), &mapConfig)
	if err != nil {
		return err
	}

	if err := mapConfig.Validate(); err != nil {
		return err
	}

	c.IndexMapping = &mapConfig
	return nil
}
