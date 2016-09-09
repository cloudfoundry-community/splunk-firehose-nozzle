package main

import (
	"flag"

	"code.cloudfoundry.org/cflager"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/cf-platform-eng/splunk-firehose-nozzle/auth"
)

var (
	skipSSL = kingpin.Flag("skip-ssl-validation", "Skip cert validation (for dev environments").
		OverrideDefaultFromEnvar("SKIP_SSL_VALIDATION").Default("false").Bool()

	uaaEndpoint = kingpin.Flag("uaa-endpoint", "UAA endpoint address").
			OverrideDefaultFromEnvar("UAA_ENDPOINT").Required().String()

	adminUaaUser = kingpin.Flag("admin-uaa-user", "Admin user.").
			OverrideDefaultFromEnvar("ADMIN_UAA_USER").Required().String()
	adminUaaSecret = kingpin.Flag("admin-uaa-secret", "Admin password.").
			OverrideDefaultFromEnvar("ADMIN_UAA_SECRET").Required().String()

	firehoseUaaUser = kingpin.Flag("firehose-uaa-user", "Firehose user.").
			OverrideDefaultFromEnvar("FIREHOSE_UAA_USER").Required().String()
	firehoseUaaSecret = kingpin.Flag("firehose-uaa-secret", "Firehose password.").
				OverrideDefaultFromEnvar("FIREHOSE_UAA_SECRET").Required().String()
)

var version = "0.0.1"

func main() {
	cflager.AddFlags(flag.CommandLine)
	flag.Parse()

	logger, _ := cflager.New("uaa-provisioner-logger")
	logger.Info("Running splunk-firehose-nozzle")

	kingpin.Version(version)
	kingpin.Parse()

	tokenRefresher := auth.NewUAATokenFetcher(*uaaEndpoint, *adminUaaUser, *adminUaaSecret, *skipSSL)

	registrar, err := auth.NewUaaRegistrar(*uaaEndpoint, tokenRefresher, *skipSSL, logger)
	if err != nil {
		logger.Fatal("Unable to initialize registrar", err)
	}

	err = registrar.RegisterFirehoseClient(*firehoseUaaUser, *firehoseUaaSecret)
	if err != nil {
		logger.Fatal("Unable to create client", err)
	}

	id, err := registrar.RegisterUser(*firehoseUaaUser, *firehoseUaaSecret)
	if err != nil {
		logger.Fatal("Unable to create user", err)
	}

	err = registrar.AddUserToGroup(id, auth.GROUP_CLOUD_CONTROLLER_ADMIN)
	if err != nil {
		logger.Fatal("Unable to add user to admin group", err)
	}
}
