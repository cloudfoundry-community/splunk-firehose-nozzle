package events_test

import (
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	. "github.com/cloudfoundry-community/splunk-firehose-nozzle/cache/cachefakes"
	fevents "github.com/cloudfoundry-community/splunk-firehose-nozzle/events"
	. "github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events", func() {
	var fcache *FakeCache
	var event *fevents.Event
	var msg *Envelope
	BeforeEach(func() {
		fcache = new(FakeCache)
		msg = CreateLogMessage()
		event = fevents.LogMessage(msg)
		event.AnnotateWithEnveloppeData(msg)
	})

	Context("given a envelope", func() {
		It("should give us what we want", func() {
			Expect(event.Fields["origin"]).To(Equal("yomomma__0"))
			Expect(event.Fields["cf_app_id"]).To(Equal("eea38ba5-53a5-4173-9617-b442d35ec2fd"))
			Expect(event.Fields["timestamp"]).To(Equal(int64(1)))
			Expect(event.Fields["source_type"]).To(Equal("Kehe"))
			Expect(event.Fields["message_type"]).To(Equal("OUT"))
			Expect(event.Fields["source_instance"]).To(Equal(">9000"))
			Expect(event.Msg).To(Equal("Help, I'm a rock! Help, I'm a rock! Help, I'm a cop! Help, I'm a cop!"))
		})
	})
	Context("given metadata", func() {
		It("Should give us the right metadata", func() {
			event.AnnotateWithMetaData(map[string]string{"extra": "field"})
			Expect(event.Fields["cf_origin"]).To(Equal("firehose"))
			Expect(event.Fields["event_type"]).To(Equal(event.Type))
			Expect(event.Fields["extra"]).To(Equal("field"))

		})

	})

	Context("given Application Metadata", func() {
		It("Should give us the right Application metadata", func() {
			fcache.GetAppStub = func(appid string) (*App, error) {
				Expect(appid).To(Equal("eea38ba5-53a5-4173-9617-b442d35ec2fd"))
				return &App{
					Name:       "App-Name",
					Guid:       appid,
					SpaceName:  "Space-Name",
					SpaceGuid:  "Space-Guid",
					OrgName:    "Org-Name",
					OrgGuid:    "Org-Guid",
					IgnoredApp: true,
				}, nil
			}
			event.AnnotateWithAppData(fcache)
			Expect(event.Fields["cf_app_name"]).To(Equal("App-Name"))
			Expect(event.Fields["cf_space_id"]).To(Equal("Space-Guid"))
			Expect(event.Fields["cf_space_name"]).To(Equal("Space-Name"))
			Expect(event.Fields["cf_org_id"]).To(Equal("Org-Guid"))
			Expect(event.Fields["cf_org_name"]).To(Equal("Org-Name"))
			Expect(event.Fields["cf_ignored_app"]).To(Equal(true))
		})
	})

	Context("ParseSelectedEvents, empty select events passed in", func() {
		It("should return a hash of only the default event", func() {
			results, err := fevents.ParseSelectedEvents("")
			Ω(err).ShouldNot(HaveOccurred())
			expected := map[string]bool{"LogMessage": true}
			Expect(results).To(Equal(expected))
		})
	})

	Context("ParseSelectedEvents, bogus event names", func() {
		It("should err out", func() {
			_, err := fevents.ParseSelectedEvents("bogus, invalid")
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("ParseSelectedEvents, valid event names", func() {
		It("should return a hash of events", func() {
			expected := map[string]bool{
				"HttpStartStop": true,
				"CounterEvent":  true,
			}
			results, err := fevents.ParseSelectedEvents("HttpStartStop,CounterEvent")
			Ω(err).ShouldNot(HaveOccurred())
			Expect(results).To(Equal(expected))
		})
	})

	Context("AuthorizedEvents", func() {
		It("should return right list of authorized events", func() {
			Expect(fevents.AuthorizedEvents()).To(Equal("ContainerMetric, CounterEvent, Error, HttpStart, HttpStartStop, HttpStop, LogMessage, ValueMetric"))
		})
	})

	Describe("ParseExtraFields", func() {
		Context("called with a empty string", func() {
			It("should return a empty hash", func() {
				expected := map[string]string{}
				Expect(fevents.ParseExtraFields("")).To(Equal(expected))
			})
		})

		Context("called with extra events", func() {
			It("should return a hash with the events we want", func() {
				expected := map[string]string{"env": "dev", "kehe": "wakawaka"}
				extraEvents := "env:dev,kehe:wakawaka"
				Expect(fevents.ParseExtraFields(extraEvents)).To(Equal(expected))
			})
		})

		Context("called with extra events with weird whitespace", func() {
			It("should return a hash with the events we want", func() {
				expected := map[string]string{"env": "dev", "kehe": "wakawaka"}
				extraEvents := "    env:      \ndev,      kehe:wakawaka   "
				Expect(fevents.ParseExtraFields(extraEvents)).To(Equal(expected))
			})
		})

		Context("called with extra events with to many values to a kv pair", func() {
			It("should return a error", func() {
				extraEvents := "to:many:values"
				_, err := fevents.ParseExtraFields(extraEvents)
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
