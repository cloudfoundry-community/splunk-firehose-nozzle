package main

import (
	"fmt"
	"os"
	"time"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/caching"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	skipSSL := kingpin.Flag("skip-ssl-validation", "Skip cert validation (for dev environments").
		OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Default("false").Bool()
	apiEndpoint := kingpin.Flag("api-endpoint", "API endpoint address").
		OverrideDefaultFromEnvar("API_ENDPOINT").Required().String()
	user := kingpin.Flag("user", "Admin user.").
		OverrideDefaultFromEnvar("API_USER").Required().String()
	password := kingpin.Flag("password", "Admin password.").
		OverrideDefaultFromEnvar("API_PASSWORD").Required().String()
	boltdbPath := kingpin.Flag("boltdb-path", "Bolt Database path ").
		Default("appinfo-bolt.db").OverrideDefaultFromEnvar("BOLTDB_PATH").String()
	kingpin.Parse()

	cfConfig := cfclient.Config{
		ApiAddress:        *apiEndpoint,
		Username:          *user,
		Password:          *password,
		SkipSslValidation: *skipSSL,
	}

	cfClient, err := cfclient.NewClient(&cfConfig)
	if err != nil {
		fmt.Printf("failed to create PCF client, error=%+v\n", err)
		os.Exit(1)
	}

	config := caching.CachingBoltConfig{
		Path: *boltdbPath,
	}
	bolt, err := caching.NewCachingBolt(cfClient, &config)
	if err != nil {
		fmt.Printf("failed to create boltdb caching client, error=%+v\n", err)
		os.Exit(1)
	}

	start := time.Now().Unix()
	fmt.Printf("Start populating boltdb=%s with app info\n", *boltdbPath)
	err = bolt.Open()
	if err != nil {
		fmt.Printf("failed to open boltdb caching, error=%+v\n", err)
		os.Exit(1)
	}
	defer bolt.Close()

	end := time.Now().Unix()
	fmt.Printf("Finish populating boltdb=%s with app info, took=%d seconds\n", *boltdbPath, end-start)
}
