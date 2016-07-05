package splunk_test

import (
	"os"

	. "github.com/cf-platform-eng/splunk-firehose-nozzle/splunk"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("config", func() {
	var (
		splunkTokenValue = "83e9f712-9734-4a48-927e-3a195b9a6beb"
		splunkHostValue  = "https://splunk.cloud.example.com"
	)

	BeforeEach(func() {
		os.Setenv("NOZZLE_SPLUNK_TOKEN", splunkTokenValue)
		os.Setenv("NOZZLE_SPLUNK_HOST", splunkHostValue)
	})

	It("returns config when all values present", func() {
		_, err := Parse()

		Expect(err).To(BeNil())
	})

	It("pulls token fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.SplunkToken).To(Equal(splunkTokenValue))
	})

	It("pulls host fron env", func() {
		config, err := Parse()

		Expect(err).To(BeNil())
		Expect(config.SplunkHost).To(Equal(splunkHostValue))
	})

	It("errors when token missing", func() {
		os.Setenv("NOZZLE_SPLUNK_TOKEN", "")

		_, err := Parse()

		Expect(err).NotTo(BeNil())
	})

	It("errors when host missing", func() {
		os.Setenv("NOZZLE_SPLUNK_HOST", "")

		_, err := Parse()

		Expect(err).NotTo(BeNil())
	})
})
