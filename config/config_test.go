package config_test

import (
	"os"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		uaaUrlValue   = "http://uaa.example.com"
		usernameValue = "user"
		passwordValue = "password"
		trafficControllerURLValue = "wss://doppler.example.com"
		firehoseSubscriptionIDValue = "splunk-nozzle-subscription"
	)

	BeforeEach(func() {
		os.Setenv("NOZZLE_UAA_URL", uaaUrlValue)
		os.Setenv("NOZZLE_USERNAME", usernameValue)
		os.Setenv("NOZZLE_PASSWORD", passwordValue)
		os.Setenv("NOZZLE_TRAFFIC_CONTROLLER_URL", trafficControllerURLValue)
		os.Setenv("NOZZLE_FIREHOSE_SUBSCRIPTION_ID", firehoseSubscriptionIDValue)
	})

	It("returns config when all values present", func() {
		_, err := Parse()

		Expect(err).To(BeNil())
	})

	DescribeTable("error on missing values", func(envName string) {
		os.Setenv(envName, "")

		_, err := Parse()

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring(envName))
	},
		Entry("uaaUrl", "NOZZLE_UAA_URL"),
		Entry("username", "NOZZLE_USERNAME"),
		Entry("password", "NOZZLE_PASSWORD"),
		Entry("trafficControllerUrl", "NOZZLE_TRAFFIC_CONTROLLER_URL"),
		Entry("FirehoseSubscriptionID", "NOZZLE_FIREHOSE_SUBSCRIPTION_ID"),
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
})
