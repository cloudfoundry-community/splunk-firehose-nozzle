package splunk

import (
	"errors"
	"fmt"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/config"
)

type SplunkConfig struct {
	SplunkToken string
	SplunkHost  string
}

func Parse() (*SplunkConfig, error) {
	splunkConfig := &SplunkConfig{}

	config.SetFromStringEnv("NOZZLE_SPLUNK_TOKEN", &splunkConfig.SplunkToken)
	if splunkConfig.SplunkToken == "" {
		return nil, errors.New(fmt.Sprintf("[%s] is required", "SplunkToken"))
	}

	config.SetFromStringEnv("NOZZLE_SPLUNK_HOST", &splunkConfig.SplunkHost)
	if splunkConfig.SplunkHost == "" {
		return nil, errors.New(fmt.Sprintf("[%s] is required", "SplunkHost"))
	}

	return splunkConfig, nil
}
