package config_test

import (
	"os"

	"github.com/cloudfoundry/sonde-go/events"

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
		firehoseSubscriptionIDValue = "nozzle-subscription"
	)

	BeforeEach(func() {
		os.Setenv("NOZZLE_UAA_URL", uaaUrlValue)
		os.Setenv("NOZZLE_USERNAME", usernameValue)
		os.Setenv("NOZZLE_PASSWORD", passwordValue)
		os.Setenv("NOZZLE_TRAFFIC_CONTROLLER_URL", trafficControllerURLValue)
		os.Setenv("NOZZLE_FIREHOSE_SUBSCRIPTION_ID", firehoseSubscriptionIDValue)
		os.Setenv("NOZZLE_INSECURE_SKIP_VERIFY", "true")
	})

	It("returns config when all values present", func() {
		_, err := Parse()

		Expect(err).To(BeNil())
	})

	DescribeTable("error on missing required values", func(envName string) {
		os.Unsetenv(envName)

		_, err := Parse()

		Expect(err).NotTo(BeNil())
		Expect(err.Error()).To(ContainSubstring(envName))
	},
		Entry("uaaUrl", "NOZZLE_UAA_URL"),
		Entry("username", "NOZZLE_USERNAME"),
		Entry("password", "NOZZLE_PASSWORD"),
		Entry("trafficControllerUrl", "NOZZLE_TRAFFIC_CONTROLLER_URL"),
		Entry("firehoseSubscriptionID", "NOZZLE_FIREHOSE_SUBSCRIPTION_ID"),
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
		os.Unsetenv("NOZZLE_INSECURE_SKIP_VERIFY")
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.InsecureSkipVerify).To(BeFalse())
	})

	Context("selected events", func() {
		It("defaults", func() {
			config, err := Parse()

			Expect(err).To(BeNil())
			Expect(config.SelectedEvents).To(HaveLen(2))
			Expect(config.SelectedEvents).To(ContainElement(events.Envelope_ValueMetric))
			Expect(config.SelectedEvents).To(ContainElement(events.Envelope_CounterEvent))
		})

		It("pulls selectedEvents from env (single value)", func() {
			os.Setenv("NOZZLE_SELECTED_EVENTS", "HttpStartStop")

			config, err := Parse()

			Expect(err).To(BeNil())
			Expect(config.SelectedEvents).To(HaveLen(1))
			Expect(config.SelectedEvents).To(ContainElement(events.Envelope_HttpStartStop))
		})

		It("pulls selectedEvents from env (multiple values)", func() {
			os.Setenv("NOZZLE_SELECTED_EVENTS", "CounterEvent, HttpStartStop,ContainerMetric")

			config, err := Parse()

			Expect(err).To(BeNil())
			Expect(config.SelectedEvents).To(HaveLen(3))
			Expect(config.SelectedEvents).To(ContainElement(events.Envelope_CounterEvent))
			Expect(config.SelectedEvents).To(ContainElement(events.Envelope_HttpStartStop))
			Expect(config.SelectedEvents).To(ContainElement(events.Envelope_ContainerMetric))
		})

		It("errors on bad value", func() {
			os.Setenv("NOZZLE_SELECTED_EVENTS", "Foo")

			_, err := Parse()

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(ContainSubstring("Foo"))
		})
	})
})
