package drain_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/drain"
)

var _ = Describe("IndexRouting", func() {

	var (
		mapping         = `{"default_index":"otherindex","mappings":{"cf_org_name":[{"value":"cf_org_name.*","index":"cf_org_name_idx"}],"cf_space_name":[{"value":"cf_space_name.*","index":"cf_space_name_idx"}],"cf_app_name":[{"value":"cf_app_name.*","index":"cf_app_name_idx"}]}}`
		invalidMapping  = `{"default_index":"","mappings":null}`
		invalidMapping2 = `{"default_index":"main","mappings":{"cf_org_name":[{"value": "*sales", "index": null}]}}`
		nilMapping      = `{"default_index":"main","mappings":{"cf_org_name":null}}`
	)

	It("Valid index routing", func() {
		var config drain.IndexMapConfig
		err := json.Unmarshal([]byte(mapping), &config)
		Ω(err).ShouldNot(HaveOccurred())

		err = config.Validate()
		Ω(err).ShouldNot(HaveOccurred())

		routing := drain.NewIndexRouting(&config)

		fields := map[string]interface{}{
			"cf_app_name": "cf_app_name_test",
		}
		idx := routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_app_name_idx"))

		fields["cf_space_name"] = "cf_space_name_test"
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_app_name_idx"))

		delete(fields, "cf_app_name")
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_space_name_idx"))

		fields["cf_org_name"] = "cf_org_name_test"
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_space_name_idx"))

		delete(fields, "cf_space_name")
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_org_name_idx"))

		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_org_name_idx"))

		delete(fields, "cf_org_name")
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("otherindex"))
	})

	It("Invalid index routing with no index setup", func() {
		var config drain.IndexMapConfig
		err := json.Unmarshal([]byte(invalidMapping), &config)
		Ω(err).ShouldNot(HaveOccurred())

		err = config.Validate()
		Ω(err).Should(HaveOccurred())
	})

	It("Invalid index routing with incorrect regex string", func() {
		var config drain.IndexMapConfig
		err := json.Unmarshal([]byte(invalidMapping2), &config)
		Ω(err).ShouldNot(HaveOccurred())

		err = config.Validate()
		Ω(err).Should(HaveOccurred())
	})

	It("Valid index routing with nil mapping", func() {
		var config drain.IndexMapConfig
		err := json.Unmarshal([]byte(nilMapping), &config)
		Ω(err).ShouldNot(HaveOccurred())

		err = config.Validate()
		Ω(err).ShouldNot(HaveOccurred())

		routing := drain.NewIndexRouting(&config)
		fields := map[string]interface{}{
			"cf_app_name": "cf_app_name_test",
		}
		idx := routing.LookupIndex(fields)
		Expect(*idx).To(Equal("main"))
	})

	It("Valid index routing with no validating", func() {
		var config drain.IndexMapConfig
		err := json.Unmarshal([]byte(mapping), &config)
		Ω(err).ShouldNot(HaveOccurred())

		routing := drain.NewIndexRouting(&config)
		fields := map[string]interface{}{
			"cf_app_name": "cf_app_name_test",
		}
		idx := routing.LookupIndex(fields)
		Expect(*idx).To(Equal("otherindex"))
	})
})
