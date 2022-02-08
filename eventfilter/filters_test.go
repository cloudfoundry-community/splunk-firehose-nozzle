package eventfilter_test

import (
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventfilter"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Rule parsing", func() {
	testError := func(filterConf string, errorMsg string) {
		filters, err := eventfilter.New(filterConf)
		Expect(filters).To(BeNil())
		Expect(err).To(MatchError(ContainSubstring(errorMsg)))
	}
	DescribeTable("throws error", testError,
		Entry("not enough fields", ":", "format must be"),
		Entry("too many fields", "xxx:yyy:zzz:rrrr", "format must be"),
		Entry("invalid value", "xxx::", "filter value must not be empty"),
		Entry("invalid filter", "xxx:yyy:zzz", "filter key must be one of"),
		Entry("invalid field", "notValid:mustContain:zzz", "filter must be one of"),
	)

	testOk := func(filterConf string, length int) {
		filters, err := eventfilter.New(filterConf)
		Expect(err).NotTo(HaveOccurred())
		Expect(filters).NotTo(BeNil(), "filters have not been initialized")
		Expect(filters.Length()).To(Equal(length), "Expected %d filter rules", length)
	}
	DescribeTable("parses ok", testOk,
		Entry("no filters at all", "", 0),
		Entry("multiple empty rules", ";;;;", 0),
		Entry("filtering on deployment", "deployment:mustContain:some deployment", 1),
		Entry("accepts whitespace between rules", "   deployment:mustContain:something   ;  origin:mustContain:someOrigin ", 2),
		Entry("accepts whitespace in filter", "deployment:  mustContain  :something", 1),

		Entry("inclusion filter on deployment", "Deployment:mustContain:something", 1),
		Entry("inclusion filter on origin", "origin:mustContain:something", 1),
		Entry("inclusion filter on job", "job:mustContain:something", 1),

		Entry("exclusion filter on deployment", "Deployment:mustNotContain:something", 1),
		Entry("exclusion filter on origin", "origin:mustNotContain:something", 1),
		Entry("exclusion filter on job", "job:mustNotContain:something", 1),
	)
})

var _ = Describe("Filtering", func() {
	msg := &events.Envelope{
		Deployment: p("p-healthwatch2-123123123"),
		Origin:     p("some origin"),
		Job:        p("some job"),
	}

	test := func(filterConf string, expected bool) {
		filters, err := eventfilter.New(filterConf)
		Expect(err).NotTo(HaveOccurred())
		Expect(filters.Accepts(msg)).
			To(Equal(expected), "Expected event {%v} to be %s", msg, tern(expected, "accepted", "discarded"))
		Expect(filters).NotTo(BeNil(), "filters have not been initialized")
	}

	DescribeTable("on", test,
		Entry("empty filter conf should accept", "", true),
		Entry("matching inclusion filter should accept", "deployment:mustContain:healthwatch2", true),
		Entry("non-matching inclusion filter should discard", "deployment:mustContain:something", false),
		Entry("matching exclusion filter should discard", "deployment:mustNotContain:healthwatch2", false),
		Entry("2nd exclusion filter should discard", "deployment:mustNotContain:health ; deployment:mustNotContain:watch", false),
		Entry("3rd exclusion filter should discard",
			"deployment:mustContain:health ; job:mustNotContain:other job ; deployment:mustNotContain:watch",
			false,
		),
		Entry("many matching inclusion filters should accept",
			"deployment:mustContain:h ; deployment:mustContain:e ; deployment:mustContain:a ; deployment:mustContain:l ; deployment:mustContain:t ; deployment:mustContain:h",
			true,
		),
		Entry("many non-matching exclusion filters should accept",
			"deployment:mustNotContain:x ; deployment:mustNotContain:y ; deployment:mustNotContain:z ; deployment:mustNotContain:u ; deployment:mustNotContain:b ; deployment:mustNotContain:r",
			true,
		),
	)
})

func p(s string) *string { return &s }

func tern(b bool, t string, f string) string {
	if b {
		return t
	}

	return f
}
