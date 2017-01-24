package config_test

import (
	"os"
	"time"

	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {

	Context("config parsing", func() {
		BeforeEach(func() {
			os.Clearenv()

			os.Setenv("API_ENDPOINT", "api.bosh-lite.com")
			os.Setenv("API_USER", "admin")
			os.Setenv("API_PASSWORD", "abc123")
			os.Setenv("SPLUNK_TOKEN", "some-token")
			os.Setenv("SPLUNK_HOST", "splunk.example.com")
			os.Setenv("SPLUNK_INDEX", "splunk_index")
		})

		It("parses config from environment", func() {
			os.Setenv("DEBUG", "true")
			os.Setenv("SKIP_SSL_VALIDATION", "true")
			os.Setenv("JOB_NAME", "my-job")
			os.Setenv("JOB_INDEX", "2")
			os.Setenv("JOB_HOST", "nozzle.example.com")

			os.Setenv("ADD_APP_INFO", "true")
			os.Setenv("BOLTDB_PATH", "foo.db")

			os.Setenv("EVENTS", "LogMessage")
			os.Setenv("EXTRA_FIELDS", "foo:bar")

			os.Setenv("FIREHOSE_KEEP_ALIVE", "42s")
			os.Setenv("FIREHOSE_SUBSCRIPTION_ID", "my-nozzle")

			os.Setenv("FLUSH_INTERVAL", "43s")

			c, err := Parse()

			Expect(err).To(BeNil())

			Expect(c.Debug).To(BeTrue())
			Expect(c.SkipSSL).To(BeTrue())
			Expect(c.JobName).To(Equal("my-job"))
			Expect(c.JobIndex).To(Equal("2"))
			Expect(c.JobHost).To(Equal("nozzle.example.com"))

			Expect(c.AddAppInfo).To(BeTrue())
			Expect(c.ApiEndpoint).To(Equal("api.bosh-lite.com"))
			Expect(c.ApiUser).To(Equal("admin"))
			Expect(c.ApiPassword).To(Equal("abc123"))
			Expect(c.BoldDBPath).To(Equal("foo.db"))

			Expect(c.WantedEvents).To(Equal("LogMessage"))
			Expect(c.ExtraFields).To(Equal("foo:bar"))

			Expect(c.KeepAlive).To(Equal(42 * time.Second))
			Expect(c.SubscriptionId).To(Equal("my-nozzle"))

			Expect(c.SplunkToken).To(Equal("some-token"))
			Expect(c.SplunkHost).To(Equal("splunk.example.com"))
			Expect(c.SplunkIndex).To(Equal("splunk_index"))
			Expect(c.FlushInterval).To(Equal(43 * time.Second))
		})

		It("check defaults", func() {
			c, err := Parse()

			Expect(err).To(BeNil())

			Expect(c.Debug).To(BeFalse())
			Expect(c.SkipSSL).To(BeFalse())
			Expect(c.JobName).To(Equal("splunk-nozzle"))
			Expect(c.JobIndex).To(Equal("-1"))
			Expect(c.JobHost).To(Equal("localhost"))

			Expect(c.AddAppInfo).To(Equal(false))
			Expect(c.BoldDBPath).To(Equal("cache.db"))

			Expect(c.WantedEvents).To(Equal("ValueMetric,CounterEvent,ContainerMetric"))

			Expect(c.KeepAlive).To(Equal(25 * time.Second))
			Expect(c.SubscriptionId).To(Equal("splunk-firehose"))

			Expect(c.FlushInterval).To(Equal(5 * time.Second))
		})

		It("parses single mapping", func() {
			os.Setenv("MAPPINGS", "deployment:cf:main")

			c, err := Parse()

			Expect(err).To(BeNil())

			Expect(c.MappingList.Mappings).To(HaveLen(1))

			mapping := c.MappingList.Mappings[0]
			Expect(mapping.Key).To(Equal("deployment"))
			Expect(mapping.Value).To(Equal("cf"))
			Expect(mapping.Index).To(Equal("main"))
		})

		It("parses single mapping", func() {
			os.Setenv("MAPPINGS", "deployment:cf:main,cf_org_id:some-guid:whatever")

			c, err := Parse()

			Expect(err).To(BeNil())

			Expect(c.MappingList.Mappings).To(HaveLen(2))

			mapping := c.MappingList.Mappings[0]
			Expect(mapping.Key).To(Equal("deployment"))
			Expect(mapping.Value).To(Equal("cf"))
			Expect(mapping.Index).To(Equal("main"))

			mapping = c.MappingList.Mappings[1]
			Expect(mapping.Key).To(Equal("cf_org_id"))
			Expect(mapping.Value).To(Equal("some-guid"))
			Expect(mapping.Index).To(Equal("whatever"))
		})
	})

	Context("mappinglist ", func() {
		It("unmarshals", func() {
			mappingList := &MappingList{}

			err := mappingList.UnmarshalText([]byte("deployment:cf:main,cf_org_id:some-guid:whatever"))

			Expect(err).To(BeNil())

			Expect(mappingList.Mappings).To(HaveLen(2))

			Expect(mappingList.Mappings[0]).To(Equal(Mapping{
				Key:   "deployment",
				Value: "cf",
				Index: "main",
			}))
		})

		It("errors on bad mapping", func() {
			mappingList := &MappingList{}

			err := mappingList.UnmarshalText([]byte("deployment:cf:main,cf_org_id::whatever"))

			Expect(err).NotTo(BeNil())
		})
	})

	Context("mapping", func() {
		It("mapping unmarshals with value input", func() {
			mapping := &Mapping{}

			err := mapping.UnmarshalText([]byte("deployment:cf:main"))

			Expect(err).To(BeNil())

			Expect(mapping.Key).To(Equal("deployment"))
			Expect(mapping.Value).To(Equal("cf"))
			Expect(mapping.Index).To(Equal("main"))
		})

		It("value can containt :", func() {
			mapping := &Mapping{}

			err := mapping.UnmarshalText([]byte("cf_app_name:foo:bar:random"))

			Expect(err).To(BeNil())

			Expect(mapping.Key).To(Equal("cf_app_name"))
			Expect(mapping.Value).To(Equal("foo:bar"))
			Expect(mapping.Index).To(Equal("random"))
		})

		It("key is required", func() {
			mapping := &Mapping{}

			err := mapping.UnmarshalText([]byte(":cf:main"))

			Expect(err).NotTo(BeNil())
		})
		It("value is required", func() {
			mapping := &Mapping{}

			err := mapping.UnmarshalText([]byte("deployment::main"))

			Expect(err).NotTo(BeNil())
		})
		It("index is required", func() {
			mapping := &Mapping{}

			err := mapping.UnmarshalText([]byte("deployment:cf:"))

			Expect(err).NotTo(BeNil())
		})
	})
})
