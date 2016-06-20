package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	UAAURL                 string
	Username               string
	Password               string
	TrafficControllerURL   string
	FirehoseSubscriptionID string
	InsecureSkipVerify     bool
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
		err := getRequiredStringEnv(name, dest)
		if err != nil {
			return nil, err
		}
	}

	err := getBoolEnv("NOZZLE_INSECURE_SKIP_VERIFY", &config.InsecureSkipVerify)
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
