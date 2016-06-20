package config_test

import (
	"os"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	var (
		uaaUrlValue                 = "http://uaa.example.com"
		usernameValue               = "user"
		passwordValue               = "password"
		trafficControllerURLValue   = "wss://doppler.example.com"
		firehoseSubscriptionIDValue = "splunk-nozzle-subscription"
		splunkTokenValue            = "83e9f712-9734-4a48-927e-3a195b9a6beb"
		splunkHostValue             = "https://splunk.cloud.example.com"
	)

	BeforeEach(func() {
		os.Setenv("NOZZLE_UAA_URL", uaaUrlValue)
		os.Setenv("NOZZLE_USERNAME", usernameValue)
		os.Setenv("NOZZLE_PASSWORD", passwordValue)
		os.Setenv("NOZZLE_TRAFFIC_CONTROLLER_URL", trafficControllerURLValue)
		os.Setenv("NOZZLE_FIREHOSE_SUBSCRIPTION_ID", firehoseSubscriptionIDValue)
		os.Setenv("NOZZLE_INSECURE_SKIP_VERIFY", "true")
		os.Setenv("NOZZLE_SPLUNK_TOKEN", splunkTokenValue)
		os.Setenv("NOZZLE_SPLUNK_HOST", splunkHostValue)
	})

	It("returns config when all values present", func() {
		_, err := Parse()

		Expect(err).To(BeNil())
	})

	DescribeTable("error on missing required values", func(envName string) {
		os.Setenv(envName, "")

		_, err := Parse()

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring(envName))
	},
		Entry("uaaUrl", "NOZZLE_UAA_URL"),
		Entry("username", "NOZZLE_USERNAME"),
		Entry("password", "NOZZLE_PASSWORD"),
		Entry("trafficControllerUrl", "NOZZLE_TRAFFIC_CONTROLLER_URL"),
		Entry("firehoseSubscriptionID", "NOZZLE_FIREHOSE_SUBSCRIPTION_ID"),
		Entry("splunkToken", "NOZZLE_SPLUNK_TOKEN"),
		Entry("splunkHost", "NOZZLE_SPLUNK_HOST"),
	)

	It("pulls uaaUrl from env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.UAAURL).To(Equal(uaaUrlValue))
	})

	It("pulls usernaname fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.Username).To(Equal(usernameValue))
	})

	It("pulls password fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.Password).To(Equal(passwordValue))
	})

	It("pulls trafficControllerURL fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.TrafficControllerURL).To(Equal(trafficControllerURLValue))
	})

	It("pulls firehoseSubscriptionID fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.FirehoseSubscriptionID).To(Equal(firehoseSubscriptionIDValue))
	})

	It("pulls insecureSkipVerify fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.InsecureSkipVerify).To(BeTrue())
	})

	It("defaults insecureSkipVerify to false", func() {
		os.Setenv("NOZZLE_INSECURE_SKIP_VERIFY", "")
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.InsecureSkipVerify).To(BeFalse())
	})

	It("pulls splunkToken fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.SplunkToken).To(Equal(splunkTokenValue))
	})

	It("pulls splunkHost fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.SplunkHost).To(Equal(splunkHostValue))
	})
})
