package splunknozzle_test

import (
	"time"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/nozzle"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nozzle", func() {
	var (
		nozzle *SplunkFirehoseNozzle = nil

		apiEndpoint    = "localhost"
		user           = "admin"
		password       = "admin"
		splunkHost     = "localhost:8088"
		splunkToken    = "12345678"
		splunkIndex    = "main"
		jobName        = "testJob"
		jobIndex       = "0"
		jobHost        = "localhost"
		skipSSL        = true
		subscriptionID = "splunk-firehose-testing"
		keepAlive      = 120 * time.Second
		addAppInfo     = true
		boltDBPath     = "/tmp/appcache.db"
		wantedEvents   = "LogMessage"
		extraFields    = "a=b,c=d"
		flushInterval  = time.Second * 10
		queueSize      = 100
		batchSize      = 100
		retries        = 3
		hecWorkers     = 16
		version        = "1.0"
		branch         = "master"
		commit         = "farwr0q98q"
		buildos        = "macos"
		debug          = true
	)

	Context("Splunk Nozzle Setup", func() {
		BeforeEach(func() {
			nozzle = NewSplunkFirehoseNozzle(apiEndpoint, user, password, splunkHost, splunkToken, splunkIndex)

			Expect(nozzle.ApiEndpoint()).To(Equal(apiEndpoint))
			Expect(nozzle.User()).To(Equal(user))
			Expect(nozzle.Password()).To(Equal(password))
			Expect(nozzle.SplunkHost()).To(Equal(splunkHost))
			Expect(nozzle.SplunkToken()).To(Equal(splunkToken))
			Expect(nozzle.SplunkIndex()).To(Equal(splunkIndex))

		})

		It("Default params", func() {
			Expect(nozzle.JobIndex()).To(Equal("-1"))
			Expect(nozzle.JobHost()).To(Equal(""))
			Expect(nozzle.JobName()).To(Equal("splunk-nozzle"))
			Expect(nozzle.SkipSSL()).To(Equal(false))
			Expect(nozzle.SubscriptionID()).To(Equal("splunk-firehose"))
			Expect(nozzle.KeepAlive()).To(Equal(25 * time.Second))
			Expect(nozzle.AddAppInfo()).To(Equal(false))
			Expect(nozzle.BoltDBPath()).To(Equal("cache.db"))
			Expect(nozzle.WantedEvents()).To(Equal("ValueMetric,CounterEvent,ContainerMetric"))
			Expect(nozzle.ExtraFields()).To(Equal(""))
			Expect(nozzle.FlushInterval()).To(Equal(5 * time.Second))
			Expect(nozzle.QueueSize()).To(Equal(10000))
			Expect(nozzle.BatchSize()).To(Equal(1000))
			Expect(nozzle.Retries()).To(Equal(5))
			Expect(nozzle.HecWorkers()).To(Equal(8))
			Expect(nozzle.Version()).To(Equal(""))
			Expect(nozzle.Branch()).To(Equal(""))
			Expect(nozzle.Commit()).To(Equal(""))
			Expect(nozzle.BuildOS()).To(Equal(""))
			Expect(nozzle.Debug()).To(Equal(false))
		})

		It("Setup params", func() {
			nozzle.WithJobName(jobName).
				WithJobIndex(jobIndex).
				WithJobHost(jobHost).
				WithSkipSSL(skipSSL).
				WithSubscriptionID(subscriptionID).
				WithKeepAlive(keepAlive).
				WithAddAppInfo(addAppInfo).
				WithBoltDBPath(boltDBPath).
				WithWantedEvents(wantedEvents).
				WithExtraFields(extraFields).
				WithFlushInterval(flushInterval).
				WithQueueSize(queueSize).
				WithBatchSize(batchSize).
				WithRetries(retries).
				WithHecWorkers(hecWorkers).
				WithVersion(version).
				WithBranch(branch).
				WithCommit(commit).
				WithBuildOS(buildos).
				WithDebug(debug)

			Expect(nozzle.JobIndex()).To(Equal(jobIndex))
			Expect(nozzle.JobHost()).To(Equal(jobHost))
			Expect(nozzle.JobName()).To(Equal(jobName))
			Expect(nozzle.SkipSSL()).To(Equal(skipSSL))
			Expect(nozzle.SubscriptionID()).To(Equal(subscriptionID))
			Expect(nozzle.KeepAlive()).To(Equal(keepAlive))
			Expect(nozzle.AddAppInfo()).To(Equal(addAppInfo))
			Expect(nozzle.BoltDBPath()).To(Equal(boltDBPath))
			Expect(nozzle.WantedEvents()).To(Equal(wantedEvents))
			Expect(nozzle.ExtraFields()).To(Equal(extraFields))
			Expect(nozzle.FlushInterval()).To(Equal(flushInterval))
			Expect(nozzle.QueueSize()).To(Equal(queueSize))
			Expect(nozzle.BatchSize()).To(Equal(batchSize))
			Expect(nozzle.Retries()).To(Equal(retries))
			Expect(nozzle.HecWorkers()).To(Equal(hecWorkers))
			Expect(nozzle.Version()).To(Equal(version))
			Expect(nozzle.Branch()).To(Equal(branch))
			Expect(nozzle.Commit()).To(Equal(commit))
			Expect(nozzle.BuildOS()).To(Equal(buildos))
			Expect(nozzle.Debug()).To(Equal(debug))
		})
	})
})
