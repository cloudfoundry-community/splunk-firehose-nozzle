package main

import (
	"fmt"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/splunknozzle"
	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/config"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	"os"
	"time"
)

// v3
func main() {
	skipSSL := kingpin.Flag("skip-ssl-validation", "Skip cert validation (for dev environments").
		Envar("SKIP_SSL_VALIDATION").Default("false").Bool()
	apiEndpoint := kingpin.Flag("api-endpoint", "API endpoint address").
		Envar("API_ENDPOINT").Required().String()
	user := kingpin.Flag("user", "Admin user.").
		Envar("API_USER").Required().String()
	password := kingpin.Flag("password", "Admin password.").
		Envar("API_PASSWORD").Required().String()
	boltdbPath := kingpin.Flag("boltdb-path", "Bolt Database path ").
		Default("appinfo-bolt.db").Envar("BOLTDB_PATH").String()
	cfClientVersion := kingpin.Flag("cf-client-version", "Version of used CfClient").
		Envar("CF_CLIENT_VERSION").Default("V2").String()
	kingpin.Parse()

	if *cfClientVersion == "V2" {
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

		config := cache.BoltdbConfig{
			Path: *boltdbPath,
		}
		bolt, err := cache.NewBoltdb(cfClient, &config)
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
		err = bolt.Close()
		if err != nil {
			fmt.Printf("failed to populate boltdb caching, error=%+v\n", err)
			os.Exit(1)
		}
		end := time.Now().Unix()
		apps, err := bolt.GetAllApps()
		if err != nil {
			fmt.Printf("failed to get apps from boltdb caching, error=%+v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Finish populating boltdb=%s with %d app info, took=%d seconds\n", *boltdbPath, len(apps), end-start)
	}

	if *cfClientVersion == "V3" {
		var skipSSLCF config.Option
		if *skipSSL {
			skipSSLCF = config.SkipTLSValidation()
		}
		cfConfig, err := config.New(*apiEndpoint, config.ClientCredentials(*user, *password), skipSSLCF, config.UserAgent("splunk-firehose-nozzle"))
		if err != nil {
			fmt.Printf("failed to create CF config, error=%+v\n", err)
			os.Exit(1)
		}

		cfClient, err := client.New(cfConfig)
		if err != nil {
			fmt.Printf("failed to create CF client, error=%+v\n", err)
			os.Exit(1)
		}

		nozzleCfClient := splunknozzle.NozzleCfClient(*cfClient)

		config := cache.BoltdbConfig{
			Path: *boltdbPath,
		}
		bolt, err := cache.NewBoltdb(nozzleCfClient, &config)
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

		err = bolt.Close()
		if err != nil {
			fmt.Printf("failed to populate boltdb caching, error=%+v\n", err)
			os.Exit(1)
		}

		end := time.Now().Unix()
		apps, err := bolt.GetAllApps()
		if err != nil {
			fmt.Printf("failed to get apps from boltdb caching, error=%+v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Finish populating boltdb=%s with %d app info, took=%d seconds\n", *boltdbPath, len(apps), end-start)
	}
}
