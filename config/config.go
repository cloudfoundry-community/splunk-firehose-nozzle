package config

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

var mappingPattern *regexp.Regexp

func init() {
	mappingPattern = regexp.MustCompile("(?P<key>\\w+):(?P<value>.+)->(?P<index>\\w+)")
}

type config struct {
	Debug    bool   `envconfig:"debug" default:"false"`
	SkipSSL  bool   `envconfig:"skip_ssl_validation" default:"false"`
	JobName  string `envconfig:"job_name" default:"splunk-nozzle"`
	JobIndex string `envconfig:"job_index" default:"-1"`
	JobHost  string `envconfig:"job_host" default:"localhost"`

	AddAppInfo  bool   `envconfig:"add_app_info" default:"false"`
	ApiEndpoint string `envconfig:"api_endpoint" required:"true"`
	ApiUser     string `envconfig:"api_user" required:"true"`
	ApiPassword string `envconfig:"api_password" required:"true"`
	BoldDBPath  string `envconfig:"boltdb_path" default:"cache.db"`

	WantedEvents string `envconfig:"events" default:"ValueMetric,CounterEvent,ContainerMetric"`
	ExtraFields  string `envconfig:"extra_fields"`

	KeepAlive      time.Duration
	SubscriptionId string `envconfig:"firehose_subscription_id" default:"splunk-firehose"`

	SplunkToken   string `envconfig:"splunk_token" required:"true"`
	SplunkHost    string `envconfig:"splunk_host" required:"true"`
	SplunkIndex   string `envconfig:"splunk_index" required:"true"`
	FlushInterval time.Duration

	MappingList MappingList `envconfig:"mappings"`

	KeepAliveRaw     string `envconfig:"firehose_keep_alive" default:"25s"`
	FlushIntervalRaw string `envconfig:"flush_interval" default:"5s"`
}

func Parse() (*config, error) {
	c := &config{}
	err := envconfig.Process("", c)
	if err != nil {
		return nil, err
	}

	c.KeepAlive, err = time.ParseDuration(c.KeepAliveRaw)
	if err != nil {
		return nil, err
	}

	c.FlushInterval, err = time.ParseDuration(c.FlushIntervalRaw)
	if err != nil {
		return nil, err
	}

	return c, nil
}

type MappingList struct {
	Mappings []Mapping
}

func (m *MappingList) UnmarshalText(text []byte) error {
	m.Mappings = []Mapping{}

	for _, split := range strings.Split(string(text), ",") {
		mapping := Mapping{}
		err := mapping.UnmarshalText([]byte(split))
		if err != nil {
			return err
		}
		m.Mappings = append(m.Mappings, mapping)
	}

	return nil
}

type Mapping struct {
	Key   string
	Value string
	Index string
}

func (m *Mapping) UnmarshalText(text []byte) error {
	if !mappingPattern.Match(text) {
		return errors.New(
			fmt.Sprintf("Mapping should be of the form field:value:index, got %v", string(text)),
		)
	}
	match := mappingPattern.FindStringSubmatch(string(text))
	for i, name := range mappingPattern.SubexpNames() {
		switch name {
		case "key":
			m.Key = match[i]
		case "value":
			m.Value = match[i]
		case "index":
			m.Index = match[i]
		}
	}

	return nil
}
