package splunk

import (
	"github.com/kelseyhightower/envconfig"
)

type SplunkConfig struct {
	SplunkToken string `required:"true" envconfig:"splunk_token"`
	SplunkHost  string `required:"true" envconfig:"splunk_host"`
}

func Parse() (*SplunkConfig, error) {
	splunkConfig := &SplunkConfig{}

	err := envconfig.Process("nozzle", splunkConfig)
	if err != nil {
		return nil, err
	}

	return splunkConfig, nil
}
