package config

import (
	"errors"
	"fmt"
	"os"
)

type Config struct {
	UAAURL                 string
	Username               string
	Password               string
	TrafficControllerURL   string
	FirehoseSubscriptionID string
}

func Parse() (*Config, error) {
	config := &Config{}

	envVars := map[string]*string{
		"NOZZLE_UAA_URL":                  &config.UAAURL,
		"NOZZLE_USERNAME":                 &config.Username,
		"NOZZLE_PASSWORD":                 &config.Password,
		"NOZZLE_TRAFFIC_CONTROLLER_URL":   &config.TrafficControllerURL,
		"NOZZLE_FIREHOSE_SUBSCRIPTION_ID": &config.FirehoseSubscriptionID,
	}

	for name, dest := range envVars {
		err := getRequiredEnvUrl(name, dest)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func getRequiredEnvUrl(name string, value *string) error {
	envValue := os.Getenv(name)
	if envValue == "" {
		return errors.New(fmt.Sprintf("[%s] is required", name))
	} else {
		*value = envValue
		return nil
	}
}
