package eventrouter_test

import (
	"github.com/cloudfoundry-community/splunk-firehose-nozzle/cache"
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
	var sink *devNullSink

	JustBeforeEach(func() {
		cache := nullCache{}
		sink = &devNullSink{}
		config := &eventrouter.Config{
			SelectedEvents: "LogMessage,HttpStart,HttpStop,HttpStartStop,ValueMetric,CounterEvent,Error,ContainerMetric",
		}
		var err error
		router, err = eventrouter.New(cache, sink, config)
		Expect(err).NotTo(HaveOccurred())
	})
	Measure("through-put", func(b Benchmarker) {
		runtime := b.Time("route", func() {
			err := pushMessages(router, messagesPerRun)
			Expect(err).NotTo(HaveOccurred())
			Expect(sink.msgs).To(Equal(messagesPerRun))
		})

		b.RecordValue("routed messages per micro second", float64(messagesPerRun)/runtime.Seconds()/float64(1000))
	}, runs)
})

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
