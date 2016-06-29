package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/cloudfoundry/sonde-go/events"
)

type Config struct {
	UAAURL                 string
	Username               string
	Password               string
	TrafficControllerURL   string
	FirehoseSubscriptionID string
	InsecureSkipVerify     bool
	SelectedEvents         []events.Envelope_EventType

	SplunkToken string
	SplunkHost  string
}

var defaultEvents = []events.Envelope_EventType{
	events.Envelope_ValueMetric,
	events.Envelope_CounterEvent,
}

func Parse() (*Config, error) {
	config := &Config{}

	envVars := map[string]*string{
		"NOZZLE_UAA_URL":                  &config.UAAURL,
		"NOZZLE_USERNAME":                 &config.Username,
		"NOZZLE_PASSWORD":                 &config.Password,
		"NOZZLE_TRAFFIC_CONTROLLER_URL":   &config.TrafficControllerURL,
		"NOZZLE_FIREHOSE_SUBSCRIPTION_ID": &config.FirehoseSubscriptionID,
		"NOZZLE_SPLUNK_TOKEN":             &config.SplunkToken,
		"NOZZLE_SPLUNK_HOST":              &config.SplunkHost,
	}

	for name, dest := range envVars {
		err := getRequiredStringEnv(name, dest)
		if err != nil {
			return nil, err
		}
	}

	err := getBoolEnv("NOZZLE_INSECURE_SKIP_VERIFY", &config.InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	err = parseSelectedEvents(&config.SelectedEvents)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func getRequiredStringEnv(name string, value *string) error {
	envValue := os.Getenv(name)
	if envValue == "" {
		return errors.New(fmt.Sprintf("[%s] is required", name))
	}

	*value = envValue
	return nil
}

func getBoolEnv(name string, value *bool) error {
	envValue := os.Getenv(name)
	if envValue == "" {
		return nil
	}

	parsedEnvValue, err := strconv.ParseBool(envValue)
	if err != nil {
		return err
	}

	*value = parsedEnvValue
	return nil
}

func parseSelectedEvents(value *[]events.Envelope_EventType) error {
	envValue := os.Getenv("NOZZLE_SELECTED_EVENTS")
	if envValue == "" {
		*value = defaultEvents
	} else {
		selectedEvents := []events.Envelope_EventType{}

		for _, envValueSplit := range strings.Split(envValue, ",") {
			envValueSlitTrimmed := strings.TrimSpace(envValueSplit)
			val, found := events.Envelope_EventType_value[envValueSlitTrimmed]
			if found {
				selectedEvents = append(selectedEvents, events.Envelope_EventType(val))
			} else {
				return errors.New(fmt.Sprintf("[%s] is required", envValueSlitTrimmed))
			}
		}
		*value = selectedEvents
	}

	return nil
}
