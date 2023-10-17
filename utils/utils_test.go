package utils_test

import (
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/utils"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Utils packages", func() {
	Describe("UUID Formated", func() {
		Context("Called with proper UUID", func() {
			It("Should return formated String", func() {
				uuid := &events.UUID{High: proto.Uint64(0), Low: proto.Uint64(0)}
				Expect(FormatUUID(uuid)).To(Equal(("00000000-0000-0000-0000-000000000000")))
			})

		})
	})

	Describe("Concat String ", func() {
		Context("Called with String Map", func() {
			It("Should return Concat string", func() {
				Expect(ConcatFormat([]string{"foo", "bar"})).To(Equal(("foo.bar")))
			})
			It("Should return Proper string", func() {
				Expect(ConcatFormat([]string{"foo   ", "bar"})).To(Equal(("foo.bar")))
			})

		})
	})

	Context("ToJson", func() {
		var jsonArray string
		var jsonMap string
		var invalidJsonArray string
		var invalidJsonMap string
		var nonJsonStr string

		BeforeEach(func() {
			// With space prefixed and suffixed
			jsonArray = `  ["splunk", "firehose", "nozzle"]    `
			jsonMap = `   {"splunk": "firehose"}  `
			invalidJsonArray = `["splunk", "firehose", "nozzle"]]`
			invalidJsonMap = `{"splunk": "firehose"}}`
			nonJsonStr = `this is a raw log`
		})

		It("Converstion", func() {
			r := ToJson(jsonArray)
			ar, ok := r.([]interface{})
			Expect(ok).To(Equal(true))
			Expect(len(ar)).To(Equal(3))

			r = ToJson(jsonMap)
			mr, ok := r.(map[string]interface{})
			Expect(ok).To(Equal(true))
			Expect(len(mr)).To(Equal(1))

			r = ToJson(invalidJsonArray)
			ar, ok = r.([]interface{})
			Expect(ok).To(Equal(false))

			s, ok := r.(string)
			Expect(s).To(Equal(invalidJsonArray))

			r = ToJson(invalidJsonMap)
			mr, ok = r.(map[string]interface{})
			Expect(ok).To(Equal(false))

			s, ok = r.(string)
			Expect(s).To(Equal(invalidJsonMap))

			r = ToJson(nonJsonStr)
			sr, ok := r.(string)
			Expect(ok).To(Equal(true))
			Expect(sr).To(Equal(nonJsonStr))
		})
	})

	It("GetHostIPInfo", func() {
		hostname, _, err := GetHostIPInfo("localhost")
		Ω(err).ShouldNot(HaveOccurred())
		Expect(hostname).To(Equal("localhost"))

		hostname, ip, err := GetHostIPInfo("invalid")
		Ω(err).Should(HaveOccurred())
		Expect(hostname).To(Equal("invalid"))
		Expect(ip).To(Equal(""))

		GetHostIPInfo("")
	})

	It("NanoSecondsToSeconds", func() {
		nano := NanoSecondsToSeconds(1501981978112315664)
		Expect(nano).To(Equal("1501981978.112315664"))
	})
})
