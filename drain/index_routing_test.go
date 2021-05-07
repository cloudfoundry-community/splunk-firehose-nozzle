package drain_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/drain"
)

var _ = Describe("IndexRouting", func() {

	var (
		mapping = `{"default_index":"otherindex","mappings":[{"by":"cf_org_id","value":"cf_org_id_test","index":"cf_org_id_idx"},{"by":"cf_space_id","value":"cf_space_id_test","index":"cf_space_id_idx"},{"by": "cf_app_id","value":"cf_app_id_test","index":"cf_app_id_idx"}]}`

		mappingWithoutNeedingAppInfo = `{"default_index":"main","mappings":[{"by":"cf_app_id","index":"cf_app_id_idx","value":"cf_app_id_test"}]}`
	)

	It("Valid index routing", func() {
		var config drain.IndexMapConfig
		err := json.Unmarshal([]byte(mapping), &config)
		Ω(err).ShouldNot(HaveOccurred())

		res := config.NeedsAppInfo(true)
		Expect(res).To(Equal(true))

		routing := drain.NewIndexRouting(&config)

		fields := map[string]interface{}{
			"cf_app_id":   "cf_app_id_test",
			"cf_space_id": "cf_space_id_test",
		}
		idx := routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_space_id_idx"))

		fields["cf_org_id"] = "cf_org_id_test"
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_org_id_idx"))

		delete(fields, "cf_org_id")
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_space_id_idx"))

		delete(fields, "cf_space_id")
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_app_id_idx"))

		delete(fields, "cf_app_id")
		idx = routing.LookupIndex(fields)
		Expect(*idx).To(Equal("otherindex"))
	})

	It("Valid index routing with removing app info", func() {
		var config drain.IndexMapConfig
		err := json.Unmarshal([]byte(mapping), &config)
		Ω(err).ShouldNot(HaveOccurred())

		res := config.NeedsAppInfo(false)
		Expect(res).To(Equal(true))

		routing := drain.NewIndexRouting(&config)

		fields := map[string]interface{}{
			"cf_app_id":   "cf_app_id_test",
			"cf_space_id": "cf_space_id_test",
			"cf_org_id":   "cf_org_id_test",
		}
		idx := routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_org_id_idx"))

		// Clean up app info
		Expect(len(fields)).To(Equal(1))
	})

	It("Mapping without needing app info", func() {
		var config drain.IndexMapConfig
		err := json.Unmarshal([]byte(mappingWithoutNeedingAppInfo), &config)
		Ω(err).ShouldNot(HaveOccurred())

		res := config.NeedsAppInfo(false)
		Expect(res).To(Equal(false))

		routing := drain.NewIndexRouting(&config)
		fields := map[string]interface{}{
			"cf_app_id": "cf_app_id_test",
		}
		idx := routing.LookupIndex(fields)
		Expect(*idx).To(Equal("cf_app_id_idx"))
	})
})
