package eventrouter_test

import (
	"fmt"

	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventfilter"
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/eventrouter"
	"github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	messagesPerRun = 100 * 1000
	runs           = 10
)

var _ = Describe("eventrouter", func() {
	var router eventrouter.Router
	var filters eventfilter.Filters
	var sink *devNullSink

	measureIt := func(f string, expectedMsgs int) {
		title := fmt.Sprintf("message through-put for filter %q", f)
		Measure(title, func(b Benchmarker) {
			runtime := b.Time("route", func() {
				err := pushMessages(router, messagesPerRun)
				Expect(err).NotTo(HaveOccurred())
				Expect(sink.msgs).To(Equal(expectedMsgs))
			})

			b.RecordValue("routed messages per micro second", float64(messagesPerRun)/runtime.Seconds()/float64(1000))
		}, runs)
	}

	JustBeforeEach(func() {
		cache := nullCache{}
		sink = &devNullSink{}
		config := &eventrouter.Config{
			SelectedEvents: "LogMessage,HttpStart,HttpStop,HttpStartStop,ValueMetric,CounterEvent,Error,ContainerMetric",
		}
		var err error
		router, err = eventrouter.New(cache, sink, config, filters)
		Expect(err).NotTo(HaveOccurred())
	})
	measureIt("", messagesPerRun)

	Describe("single accepting filter", func() {
		var filterRules = "deployment:mustContain:some deployment"
		BeforeEach(func() {
			var err error
			filters, err = eventfilter.New(filterRules)
			Expect(err).NotTo(HaveOccurred())
		})
		measureIt(filterRules, messagesPerRun)
	})

	Describe("multiple filters, first one accepts, others don't match", func() {
		filterRules := "deployment:mustContain:ome depl " + manyNotMatchingThings
		BeforeEach(func() {
			var err error
			filters, err = eventfilter.New(filterRules)
			Expect(err).NotTo(HaveOccurred())
		})
		measureIt(first(filterRules, 40), messagesPerRun)
	})

	Describe("multiple filters, last one accepts", func() {
		filterRules := manyNotMatchingThings + "  deployment:mustContain:ome depl"
		BeforeEach(func() {
			var err error
			filters, err = eventfilter.New(filterRules)
			Expect(err).NotTo(HaveOccurred())
		})
		measureIt(last(filterRules, 40), messagesPerRun)
	})

	Describe("multiple filters, first one discards", func() {
		filterRules := "deployment:mustNotContain:ome depl " + manyNotMatchingThings
		BeforeEach(func() {
			var err error
			filters, err = eventfilter.New(filterRules)
			Expect(err).NotTo(HaveOccurred())
		})
		measureIt(first(filterRules, 40), 0)
	})

	Describe("multiple filters, first one accepts, second one discards, others don't match", func() {
		filterRules := "deployment:mustContain:ome depl; deployment:mustNotContain:some " + manyNotMatchingThings
		BeforeEach(func() {
			var err error
			filters, err = eventfilter.New(filterRules)
			Expect(err).NotTo(HaveOccurred())
		})
		measureIt(first(filterRules, 60), 0)
	})
})

// manyNotMatchingThings is used to create some closures that get passed into
// the router as filters, to be able to measure if and how the number of those
// closures influences the router's performance.
const manyNotMatchingThings = ";" +
	"origin:mustNotContain:ignore0; origin:mustNotContain:ignore1; origin:mustNotContain:ignore2; origin:mustNotContain:ignore3;" +
	"origin:mustNotContain:ignore4; origin:mustNotContain:ignore5; origin:mustNotContain:ignore6; origin:mustNotContain:ignore7;" +
	"origin:mustNotContain:ignore8; origin:mustNotContain:ignore9; origin:mustNotContain:ignore10; origin:mustNotContain:ignore11;" +
	"origin:mustNotContain:ignore12; origin:mustNotContain:ignore13; origin:mustNotContain:ignore14; origin:mustNotContain:ignore15;" +
	"origin:mustNotContain:ignore16; origin:mustNotContain:ignore17; origin:mustNotContain:ignore18; origin:mustNotContain:ignore19;"

func pushMessages(r eventrouter.Router, nrOfMsgs int) error {
	events := make(chan *events.Envelope, 10)
	stop := make(chan struct{})

	defer func() { close(stop) }()

	go eventGenerator(nrOfMsgs, events, stop)

	for msg := range events {
		err := r.Route(msg)
		if err != nil {
			return err
		}
	}

	return nil
}

func eventGenerator(count int, ch chan *events.Envelope, stop chan struct{}) {
	defer func() { close(ch) }()

	for i := 0; i < count; i++ {
		select {
		case <-stop:
			return
		default:
			ch <- &events.Envelope{
				Deployment: p("some deployment"),
				Origin:     p("some origin"),
			}
		}
	}
}

func p(s string) *string {
	return &s
}

type devNullSink struct {
	msgs int
}

func (s *devNullSink) Open() error  { return nil }
func (s *devNullSink) Close() error { return nil }
func (s *devNullSink) Write(fields map[string]interface{}, msg string) error {
	s.msgs += 1
	return nil
}

type nullCache struct{}

func (nullCache) Open() error                                { return nil }
func (nullCache) Close() error                               { return nil }
func (nullCache) GetAllApps() (map[string]*cache.App, error) { return map[string]*cache.App{}, nil }
func (nullCache) GetApp(string) (*cache.App, error)          { return &cache.App{}, nil }

func first(s string, l int) string {
	return s[0:l] + "[...]"
}

func last(s string, l int) string {
	return "[...]" + s[len(s)-l:]
}
