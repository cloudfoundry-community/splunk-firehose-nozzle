package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	UAAURL                 string `required:"true" envconfig:"uaa_url"`
	Username               string `required:"true"`
	Password               string `required:"true"`
	TrafficControllerURL   string `required:"true" envconfig:"traffic_controller_url"`
	FirehoseSubscriptionID string `required:"true" envconfig:"firehose_subscription_id"`
	InsecureSkipVerify     bool   `default:"false" envconfig:"insecure_skip_verify"`

	SelectedEvents []events.Envelope_EventType `ignored:"true"`
}

var defaultEvents = []events.Envelope_EventType{
	events.Envelope_ValueMetric,
	events.Envelope_CounterEvent,
}

func Parse() (*Config, error) {
	config := &Config{}

	err := envconfig.Process("nozzle", config)
	if err != nil {
		return nil, err
	}

	selectedEvents, err := parseSelectedEvents()
	if err != nil {
		return nil, err
	}
	config.SelectedEvents = selectedEvents

	return config, nil
}

func parseSelectedEvents() ([]events.Envelope_EventType, error) {
	envValue := os.Getenv("NOZZLE_SELECTED_EVENTS")
	if envValue == "" {
		return defaultEvents, nil
	} else {
		selectedEvents := []events.Envelope_EventType{}

		for _, envValueSplit := range strings.Split(envValue, ",") {
			envValueSlitTrimmed := strings.TrimSpace(envValueSplit)
			val, found := events.Envelope_EventType_value[envValueSlitTrimmed]
			if found {
				selectedEvents = append(selectedEvents, events.Envelope_EventType(val))
			} else {
				return nil, errors.New(fmt.Sprintf("[%s] is not a valid event type", envValueSlitTrimmed))
			}
		}
		return selectedEvents, nil
	}
}
